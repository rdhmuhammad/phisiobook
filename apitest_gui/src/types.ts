export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH' | 'HEAD' | 'OPTIONS';
export type BodyType = 'none' | 'raw' | 'multipart' | 'urlencoded';
export type RequestType = 'rest' | 'socketio' | 'websocket';

export interface FormField { id: string; key: string; value: string; enabled: boolean; }

export interface RequestBody {
  mode: BodyType;
  raw: string;
  multipart: FormField[];
  urlencoded: FormField[];
}

export interface SocketIoEvent {
  id: string;
  name: string;
  payload: string;
  useAck: boolean;
}

export interface SocketIoConfig {
  url: string;
  namespace: string;
  events: SocketIoEvent[];
}

export interface ApiRequest {
  id: string;
  name: string;
  folderId: string | null;
  fromPostman: boolean;
  type: RequestType;
  method: HttpMethod;
  url: string;
  headers: { key: string; value: string }[];
  body: RequestBody;
  pathVars: Record<string, string>;
  queryParams: FormField[];
  socketio: SocketIoConfig;
}

export interface Folder { id: string; name: string; isOpen: boolean; }
