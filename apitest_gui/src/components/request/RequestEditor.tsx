import { useState } from 'react';
import { Send } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from '@/components/ui/select';
import { BodyEditor } from './BodyEditor';
import { SocketIoEditor } from './SocketIoEditor';
import { WebSocketEditor } from './WebSocketEditor';
import { cn } from '@/lib/utils';
import type { ApiRequest, HttpMethod, RequestType, FormField } from '@/types';

const METHODS: HttpMethod[] = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS'];

const METHOD_COLORS: Record<HttpMethod, string> = {
  GET: 'text-emerald-600',
  POST: 'text-sky-600',
  PUT: 'text-amber-600',
  PATCH: 'text-violet-600',
  DELETE: 'text-rose-600',
  HEAD: 'text-slate-500',
  OPTIONS: 'text-slate-500',
};

const TYPE_OPTIONS: { value: RequestType; label: string }[] = [
  { value: 'rest', label: 'REST' },
  { value: 'socketio', label: 'Socket.IO' },
  { value: 'websocket', label: 'WebSocket' },
];

interface ResponseData {
  status: number;
  statusText: string;
  headers: Record<string, string>;
  body: string;
  time: number;
}

interface RequestEditorProps {
  request: ApiRequest;
  onChange: (updated: ApiRequest) => void;
  envVars: Record<string, string>;
}

function resolveUrl(
  url: string,
  envVars: Record<string, string>,
  pathVars: Record<string, string>,
): string {
  return url
    .replace(/\{\{(\w+)\}\}/g, (_, k) => envVars[k] ?? `{{${k}}}`)
    .replace(/:(\w+)/g, (_, k) => pathVars[k] || `:${k}`);
}

function buildQueryString(params: FormField[]): string {
  const enabled = params.filter((p) => p.enabled && p.key);
  if (enabled.length === 0) return '';
  return '?' + enabled.map((p) => `${encodeURIComponent(p.key)}=${encodeURIComponent(p.value)}`).join('&');
}

function StatusChip({ status }: { status: number }) {
  const color =
    status <= 299 ? 'bg-emerald-100 text-emerald-700' :
    status <= 399 ? 'bg-amber-100 text-amber-700' :
    'bg-rose-100 text-rose-700';
  return (
    <span className={cn('text-xs font-bold px-2 py-0.5 rounded', color)}>
      {status}
    </span>
  );
}

function ResponsePanel({ response }: { response: ResponseData }) {
  let prettyBody = response.body;
  try {
    prettyBody = JSON.stringify(JSON.parse(response.body), null, 2);
  } catch { /* raw */ }

  return (
    <div className="border-t border-border flex flex-col" style={{ height: '40%' }}>
      <div className="flex items-center gap-2 px-4 py-2 shrink-0">
        <StatusChip status={response.status} />
        <span className="text-xs text-muted-foreground">{response.statusText}</span>
        <span className="ml-auto text-xs text-muted-foreground bg-muted px-2 py-0.5 rounded">
          {response.time}ms
        </span>
      </div>
      <Tabs defaultValue="body" className="flex flex-col flex-1 overflow-hidden">
        <TabsList className="shrink-0 h-7 bg-transparent border-b border-border rounded-none justify-start gap-0 px-4">
          <TabsTrigger
            value="body"
            className="rounded-none border-b-2 border-transparent data-[state=active]:border-foreground data-[state=active]:bg-transparent h-7 px-3 text-xs"
          >
            Body
          </TabsTrigger>
          <TabsTrigger
            value="headers"
            className="rounded-none border-b-2 border-transparent data-[state=active]:border-foreground data-[state=active]:bg-transparent h-7 px-3 text-xs"
          >
            Headers
          </TabsTrigger>
        </TabsList>
        <TabsContent value="body" className="flex-1 overflow-auto mt-0 p-3">
          <pre className="font-mono text-xs whitespace-pre-wrap break-all">{prettyBody}</pre>
        </TabsContent>
        <TabsContent value="headers" className="flex-1 overflow-auto mt-0 p-3">
          <div className="flex flex-col gap-1">
            {Object.entries(response.headers).map(([k, v]) => (
              <div key={k} className="flex gap-2 text-xs font-mono">
                <span className="text-muted-foreground shrink-0">{k}:</span>
                <span className="break-all">{v}</span>
              </div>
            ))}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}

function ParamsTab({
  params,
  onChange,
}: {
  params: FormField[];
  onChange: (params: FormField[]) => void;
}) {
  function addRow() {
    onChange([...params, { id: crypto.randomUUID(), key: '', value: '', enabled: true }]);
  }
  function updateRow(id: string, patch: Partial<FormField>) {
    onChange(params.map((p) => (p.id === id ? { ...p, ...patch } : p)));
  }
  function removeRow(id: string) {
    onChange(params.filter((p) => p.id !== id));
  }

  return (
    <div className="flex flex-col gap-1 p-3">
      <div className="grid grid-cols-[24px_1fr_1fr_28px] gap-1 px-1 text-xs text-muted-foreground font-medium">
        <span />
        <span>Key</span>
        <span>Value</span>
        <span />
      </div>
      {params.map((p) => (
        <div key={p.id} className="grid grid-cols-[24px_1fr_1fr_28px] gap-1 items-center">
          <input
            type="checkbox"
            checked={p.enabled}
            onChange={(e) => updateRow(p.id, { enabled: e.target.checked })}
            className="w-4 h-4 accent-primary"
          />
          <Input
            className="h-7 text-xs"
            value={p.key}
            onChange={(e) => updateRow(p.id, { key: e.target.value })}
            placeholder="key"
          />
          <Input
            className="h-7 text-xs"
            value={p.value}
            onChange={(e) => updateRow(p.id, { value: e.target.value })}
            placeholder="value"
          />
          <Button size="icon" variant="ghost" className="h-7 w-7" onClick={() => removeRow(p.id)}>
            <span className="text-muted-foreground text-xs">✕</span>
          </Button>
        </div>
      ))}
      <Button size="sm" variant="ghost" className="mt-1 h-7 text-xs self-start gap-1" onClick={addRow}>
        + Add
      </Button>
    </div>
  );
}

export function RequestEditor({ request, onChange, envVars }: RequestEditorProps) {
  const [response, setResponse] = useState<ResponseData | null>(null);
  const [sending, setSending] = useState(false);

  if (request.type === 'socketio') {
    return <SocketIoEditor request={request} onChange={onChange} />;
  }
  if (request.type === 'websocket') {
    return <WebSocketEditor request={request} onChange={onChange} />;
  }

  const resolvedBase = resolveUrl(request.url, envVars, request.pathVars);

  async function handleSend() {
    const fetchUrl = resolvedBase + buildQueryString(request.queryParams);
    setSending(true);
    setResponse(null);
    const t0 = Date.now();
    try {
      const headerObj: Record<string, string> = {};
      request.headers.forEach((h) => {
        if (h.key) headerObj[h.key] = h.value;
      });

      let fetchBody: BodyInit | undefined;
      if (request.body.mode === 'raw' && request.body.raw) {
        fetchBody = request.body.raw;
        if (!headerObj['Content-Type']) headerObj['Content-Type'] = 'application/json';
      } else if (request.body.mode === 'urlencoded') {
        const fd = new URLSearchParams();
        request.body.urlencoded.filter((f) => f.enabled && f.key).forEach((f) => fd.append(f.key, f.value));
        fetchBody = fd;
      } else if (request.body.mode === 'multipart') {
        const fd = new FormData();
        request.body.multipart.filter((f) => f.enabled && f.key).forEach((f) => fd.append(f.key, f.value));
        fetchBody = fd;
      }

      const res = await fetch(fetchUrl, {
        method: request.method,
        headers: headerObj,
        body: ['GET', 'HEAD'].includes(request.method) ? undefined : fetchBody,
      });

      const resHeaders: Record<string, string> = {};
      res.headers.forEach((v, k) => { resHeaders[k] = v; });
      const bodyText = await res.text();
      setResponse({
        status: res.status,
        statusText: res.statusText,
        headers: resHeaders,
        body: bodyText,
        time: Date.now() - t0,
      });
    } catch (err) {
      setResponse({
        status: 0,
        statusText: String(err),
        headers: {},
        body: String(err),
        time: Date.now() - t0,
      });
    } finally {
      setSending(false);
    }
  }

  return (
    <div className="flex flex-col h-full">
      {/* Name + type selector */}
      <div className="flex items-center gap-2 px-4 pt-4 pb-2">
        {request.fromPostman ? (
          <span className="text-base font-semibold flex-1 truncate">{request.name}</span>
        ) : (
          <Input
            className="flex-1 text-base font-semibold border-none shadow-none px-0 h-auto focus-visible:ring-0 bg-transparent"
            value={request.name}
            onChange={(e) => onChange({ ...request, name: e.target.value })}
            placeholder="Request name"
          />
        )}
        <Select
          value={request.type}
          onValueChange={(v) => onChange({ ...request, type: v as RequestType })}
        >
          <SelectTrigger className="w-[110px] h-7 text-xs">
            <span className="text-xs">{TYPE_OPTIONS.find((o) => o.value === request.type)?.label}</span>
          </SelectTrigger>
          <SelectContent>
            {TYPE_OPTIONS.map((o) => (
              <SelectItem key={o.value} value={o.value}>
                <span className="text-xs">{o.label}</span>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <Separator />

      {/* Method + URL + Send */}
      <div className="flex items-center gap-2 px-4 py-3">
        {request.fromPostman ? (
          <span className={cn('text-xs font-bold w-[60px] shrink-0', METHOD_COLORS[request.method])}>
            {request.method}
          </span>
        ) : (
          <Select
            value={request.method}
            onValueChange={(v) => onChange({ ...request, method: v as HttpMethod })}
          >
            <SelectTrigger className="w-[110px] h-8 text-xs font-bold">
              <span className={cn('font-bold text-xs', METHOD_COLORS[request.method])}>
                {request.method}
              </span>
            </SelectTrigger>
            <SelectContent>
              {METHODS.map((m) => (
                <SelectItem key={m} value={m}>
                  <span className={cn('font-bold text-xs', METHOD_COLORS[m])}>{m}</span>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}

        {request.fromPostman ? (
          <div className="flex-1 h-8 px-2 flex items-center bg-muted rounded text-sm font-mono overflow-x-auto whitespace-nowrap">
            {resolvedBase}
          </div>
        ) : (
          <Input
            className="flex-1 h-8 text-sm font-mono"
            value={request.url}
            onChange={(e) => onChange({ ...request, url: e.target.value })}
            placeholder="https://api.example.com/endpoint"
          />
        )}

        <Button size="sm" className="h-8 gap-1.5 shrink-0" onClick={handleSend} disabled={sending}>
          <Send className="w-3.5 h-3.5" />
          {sending ? 'Sending…' : 'Send'}
        </Button>
      </div>

      <Separator />

      {/* Tabs area */}
      <div className={cn('overflow-hidden', response ? 'flex-1' : 'flex-1')}>
        {request.fromPostman ? (
          <Tabs defaultValue="params" className="flex flex-col h-full">
            <TabsList className="shrink-0 h-8 bg-transparent border-b border-border rounded-none justify-start gap-0 px-0">
              {['params', 'path-vars', 'body', 'headers'].map((tab) => (
                <TabsTrigger
                  key={tab}
                  value={tab}
                  className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:shadow-none h-8 px-3 text-xs"
                >
                  {tab === 'params' ? 'Params' : tab === 'path-vars' ? 'Path Variables' : tab === 'body' ? 'Body' : 'Headers'}
                </TabsTrigger>
              ))}
            </TabsList>

            <TabsContent value="params" className="flex-1 overflow-auto mt-0">
              <ParamsTab
                params={request.queryParams}
                onChange={(queryParams) => onChange({ ...request, queryParams })}
              />
            </TabsContent>

            <TabsContent value="path-vars" className="flex-1 overflow-auto mt-0 p-3">
              {Object.keys(request.pathVars).length === 0 ? (
                <p className="text-xs text-muted-foreground">No path variables</p>
              ) : (
                <div className="flex flex-col gap-2">
                  {Object.entries(request.pathVars).map(([key, val]) => (
                    <div key={key} className="flex items-center gap-2">
                      <span className="text-xs font-mono text-muted-foreground w-32 shrink-0">:{key}</span>
                      <Input
                        className="h-7 text-xs"
                        value={val}
                        onChange={(e) =>
                          onChange({ ...request, pathVars: { ...request.pathVars, [key]: e.target.value } })
                        }
                        placeholder={`value for :${key}`}
                      />
                    </div>
                  ))}
                </div>
              )}
            </TabsContent>

            <TabsContent value="body" className="flex-1 overflow-auto mt-0 p-3">
              {request.body.raw === 'none' || !request.body.raw ? (
                <p className="text-xs text-muted-foreground">No body</p>
              ) : (
                <pre className="font-mono text-xs whitespace-pre-wrap break-all">{request.body.raw}</pre>
              )}
            </TabsContent>

            <TabsContent value="headers" className="flex-1 overflow-auto mt-0 p-3">
              {request.headers.length === 0 ? (
                <p className="text-xs text-muted-foreground">No headers</p>
              ) : (
                <div className="flex flex-col gap-1">
                  {request.headers.map((h, i) => (
                    <div key={i} className="flex gap-2 text-xs font-mono">
                      <span className="text-muted-foreground shrink-0">{h.key}:</span>
                      <span className="break-all">{h.value}</span>
                    </div>
                  ))}
                </div>
              )}
            </TabsContent>
          </Tabs>
        ) : (
          <BodyEditor
            body={request.body}
            onChange={(body) => onChange({ ...request, body })}
          />
        )}
      </div>

      {response && <ResponsePanel response={response} />}
    </div>
  );
}
