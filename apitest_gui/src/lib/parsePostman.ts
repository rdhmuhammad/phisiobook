import type { ApiRequest, Folder, FormField, HttpMethod, RequestBody } from '@/types';

interface PostmanVariable { key: string; value: string }
interface PostmanHeader { key: string; value: string }
interface PostmanUrlQuery { key: string; value: string; disabled?: boolean }
interface PostmanUrlVariable { key: string; value?: string }
interface PostmanUrl {
  raw: string;
  path?: string[];
  query?: PostmanUrlQuery[];
  variable?: PostmanUrlVariable[];
}
interface PostmanBody {
  mode: string;
  raw?: string;
  formdata?: { key: string; value: string; disabled?: boolean }[];
  urlencoded?: { key: string; value: string; disabled?: boolean }[];
}
interface PostmanRequest {
  method: string;
  header?: PostmanHeader[];
  url?: PostmanUrl;
  body?: PostmanBody;
}
interface PostmanSocketIoEvent {
  id?: string;
  name: string;
  payload: string;
  useAck: boolean;
}

interface PostmanSocketIoConfig {
  url: string;
  namespace: string;
  events: PostmanSocketIoEvent[];
}

interface PostmanItem {
  id?: string;
  name: string;
  request?: PostmanRequest;
  item?: PostmanItem[];
  'x-type'?: string;
  socketio?: PostmanSocketIoConfig;
}
interface PostmanCollection {
  info?: { name?: string };
  item?: PostmanItem[];
  variable?: PostmanVariable[];
}

function parseBody(postmanBody: PostmanBody | undefined): RequestBody {
  if (!postmanBody) return { type: 'none', json: '', multipart: [], urlencoded: [] };

  if (postmanBody.mode === 'raw') {
    return { type: 'json', json: postmanBody.raw ?? '', multipart: [], urlencoded: [] };
  }
  if (postmanBody.mode === 'formdata') {
    const multipart: FormField[] = (postmanBody.formdata ?? []).map((f) => ({
      id: crypto.randomUUID(),
      key: f.key,
      value: f.value,
      enabled: !f.disabled,
    }));
    return { type: 'multipart', json: '', multipart, urlencoded: [] };
  }
  if (postmanBody.mode === 'urlencoded') {
    const urlencoded: FormField[] = (postmanBody.urlencoded ?? []).map((f) => ({
      id: crypto.randomUUID(),
      key: f.key,
      value: f.value,
      enabled: !f.disabled,
    }));
    return { type: 'urlencoded', json: '', multipart: [], urlencoded };
  }
  return { type: 'none', json: '', multipart: [], urlencoded: [] };
}

function parseRequestItem(item: PostmanItem, folderId: string | null): ApiRequest {
  const id = item.id ?? crypto.randomUUID();

  // Socket.IO item — detected by presence of socketio config
  if (item.socketio) {
    const sio = item.socketio;
    return {
      id,
      name: item.name,
      folderId,
      fromPostman: true,
      type: 'socketio',
      method: 'GET',
      url: '',
      headers: [],
      body: { type: 'none', json: '', multipart: [], urlencoded: [] },
      pathVars: {},
      queryParams: [],
      socketio: {
        url: sio.url,
        namespace: sio.namespace,
        events: (sio.events ?? []).map((ev) => ({
          id: ev.id ?? crypto.randomUUID(),
          name: ev.name,
          payload: ev.payload,
          useAck: ev.useAck,
        })),
      },
    };
  }

  const req = item.request!;
  const url = req.url;
  const rawUrl = url?.raw ?? '';

  // Detect path vars from :varName segments and from url.variable[]
  const pathVarKeys = new Set<string>();
  (url?.path ?? []).forEach((segment) => {
    const m = segment.match(/^:(\w+)$/);
    if (m) pathVarKeys.add(m[1]);
  });
  (url?.variable ?? []).forEach((v) => {
    if (v.key) pathVarKeys.add(v.key);
  });

  const pathVars: Record<string, string> = {};
  pathVarKeys.forEach((k) => { pathVars[k] = ''; });

  const queryParams: FormField[] = (url?.query ?? []).map((q) => ({
    id: crypto.randomUUID(),
    key: q.key,
    value: q.value,
    enabled: !q.disabled,
  }));

  const headers = (req.header ?? []).map((h) => ({ key: h.key, value: h.value }));

  return {
    id,
    name: item.name,
    folderId,
    fromPostman: true,
    type: 'rest',
    method: (req.method ?? 'GET').toUpperCase() as HttpMethod,
    url: rawUrl,
    headers,
    body: parseBody(req.body),
    pathVars,
    queryParams,
    socketio: { url: '', namespace: '/', events: [] },
  };
}

function walkItems(
  items: PostmanItem[],
  folderId: string | null,
  folders: Folder[],
  requests: ApiRequest[],
) {
  for (const item of items) {
    if (item.item !== undefined) {
      // It's a folder (has children array, even if empty)
      const id = item.id ?? crypto.randomUUID();
      folders.push({ id, name: item.name, isOpen: true });
      walkItems(item.item, id, folders, requests);
    } else if (item.socketio || item.request) {
      requests.push(parseRequestItem(item, folderId));
    }
  }
}

export function parsePostman(json: unknown): {
  folders: Folder[];
  requests: ApiRequest[];
  envVars: Record<string, string>;
} {
  const col = json as PostmanCollection;
  const folders: Folder[] = [];
  const requests: ApiRequest[] = [];

  const envVars: Record<string, string> = {};
  (col.variable ?? []).forEach((v) => {
    envVars[v.key] = v.value;
  });

  walkItems(col.item ?? [], null, folders, requests);

  return { folders, requests, envVars };
}
