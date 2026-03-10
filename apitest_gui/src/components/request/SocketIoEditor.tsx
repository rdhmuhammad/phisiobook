import { useRef, useState, useEffect } from 'react';
import { io, Socket } from 'socket.io-client';
import { Plus, Trash2 } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';
import type { ApiRequest, SocketIoEvent, RequestType } from '@/types';

const TYPE_OPTIONS: { value: RequestType; label: string }[] = [
  { value: 'rest', label: 'REST' },
  { value: 'socketio', label: 'Socket.IO' },
  { value: 'websocket', label: 'WebSocket' },
];

type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error';

type LogDirection = 'sent' | 'received' | 'ack' | 'error';

interface LogEntry {
  id: string;
  ts: Date;
  direction: LogDirection;
  event: string;
  payload: string;
}

const STATUS_COLORS: Record<ConnectionStatus, string> = {
  disconnected: 'bg-slate-400',
  connecting: 'bg-amber-400',
  connected: 'bg-emerald-500',
  error: 'bg-rose-500',
};

const STATUS_LABELS: Record<ConnectionStatus, string> = {
  disconnected: 'Disconnected',
  connecting: 'Connecting…',
  connected: 'Connected',
  error: 'Error',
};

const LOG_COLORS: Record<LogDirection, string> = {
  sent: 'text-sky-600',
  received: 'text-emerald-600',
  ack: 'text-violet-600',
  error: 'text-rose-600',
};

const LOG_ARROWS: Record<LogDirection, string> = {
  sent: '→',
  received: '←',
  ack: '✓',
  error: '✗',
};

interface SocketIoEditorProps {
  request: ApiRequest;
  onChange: (updated: ApiRequest) => void;
}

export function SocketIoEditor({ request, onChange }: SocketIoEditorProps) {
  const socketRef = useRef<Socket | null>(null);
  const [status, setStatus] = useState<ConnectionStatus>('disconnected');
  const [log, setLog] = useState<LogEntry[]>([]);
  const { socketio } = request;

  useEffect(() => {
    return () => {
      socketRef.current?.disconnect();
    };
  }, []);

  function addLogEntry(direction: LogDirection, event: string, payload: unknown) {
    const entry: LogEntry = {
      id: crypto.randomUUID(),
      ts: new Date(),
      direction,
      event,
      payload: typeof payload === 'string' ? payload : JSON.stringify(payload),
    };
    setLog((prev) => [entry, ...prev]);
  }

  function connect() {
    if (socketRef.current) {
      socketRef.current.disconnect();
      socketRef.current = null;
    }

    setStatus('connecting');
    const fullUrl = socketio.url + socketio.namespace;
    const socket = io(fullUrl, { autoConnect: false });
    socketRef.current = socket;

    socket.on('connect', () => setStatus('connected'));
    socket.on('disconnect', () => setStatus('disconnected'));
    socket.on('connect_error', (err) => {
      setStatus('error');
      addLogEntry('error', 'connect_error', err.message);
    });
    socket.onAny((eventName: string, ...args: unknown[]) => {
      addLogEntry('received', eventName, args.length === 1 ? args[0] : args);
    });

    socket.connect();
  }

  function disconnect() {
    socketRef.current?.disconnect();
    socketRef.current = null;
    setStatus('disconnected');
  }

  function emitEvent(event: SocketIoEvent) {
    if (!socketRef.current || status !== 'connected') return;

    let parsed: unknown;
    try {
      parsed = event.payload.trim() ? JSON.parse(event.payload) : undefined;
    } catch {
      addLogEntry('error', event.name, 'Invalid JSON payload');
      return;
    }

    addLogEntry('sent', event.name, parsed ?? '');

    if (event.useAck) {
      socketRef.current.emit(event.name, parsed, (ack: unknown) => {
        addLogEntry('ack', event.name, ack);
      });
    } else {
      socketRef.current.emit(event.name, parsed);
    }
  }

  function updateSocketio(patch: Partial<typeof socketio>) {
    onChange({ ...request, socketio: { ...socketio, ...patch } });
  }

  function addEvent() {
    const newEvent: SocketIoEvent = {
      id: crypto.randomUUID(),
      name: '',
      payload: '',
      useAck: false,
    };
    updateSocketio({ events: [...socketio.events, newEvent] });
  }

  function updateEvent(id: string, patch: Partial<SocketIoEvent>) {
    updateSocketio({
      events: socketio.events.map((ev) => (ev.id === id ? { ...ev, ...patch } : ev)),
    });
  }

  function deleteEvent(id: string) {
    updateSocketio({ events: socketio.events.filter((ev) => ev.id !== id) });
  }

  function formatTime(ts: Date) {
    return ts.toTimeString().slice(0, 8);
  }

  function truncate(str: string, max = 120) {
    return str.length > max ? str.slice(0, max) + '…' : str;
  }

  return (
    <div className="flex flex-col h-full">
      {/* Request name + type selector */}
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

      {/* URL + namespace + connect controls */}
      <div className="flex items-center gap-2 px-4 py-3">
        <Input
          className="flex-1 h-8 text-sm font-mono"
          value={socketio.url}
          onChange={(e) => updateSocketio({ url: e.target.value })}
          placeholder="http://localhost:3001"
        />
        <Input
          className="w-[120px] h-8 text-sm font-mono"
          value={socketio.namespace}
          onChange={(e) => updateSocketio({ namespace: e.target.value })}
          placeholder="/namespace"
        />
        {status === 'connected' || status === 'connecting' ? (
          <Button size="sm" variant="destructive" className="h-8 shrink-0" onClick={disconnect}>
            Disconnect
          </Button>
        ) : (
          <Button size="sm" className="h-8 shrink-0" onClick={connect}>
            Connect
          </Button>
        )}
        <div className="flex items-center gap-1.5 shrink-0">
          <span className={cn('w-2 h-2 rounded-full', STATUS_COLORS[status])} />
          <span className="text-xs text-muted-foreground">{STATUS_LABELS[status]}</span>
        </div>
      </div>

      <Separator />

      {/* Tabs */}
      <div className="flex-1 overflow-hidden">
        <Tabs defaultValue="events" className="flex flex-col h-full">
          <TabsList className="mx-4 mt-2 w-auto justify-start rounded-none border-b bg-transparent p-0 h-auto gap-0">
            <TabsTrigger
              value="events"
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-foreground data-[state=active]:bg-transparent px-3 pb-2 text-xs"
            >
              Events
            </TabsTrigger>
            <TabsTrigger
              value="log"
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-foreground data-[state=active]:bg-transparent px-3 pb-2 text-xs"
            >
              Log
              {log.length > 0 && (
                <span className="ml-1.5 text-[10px] bg-muted rounded-full px-1.5 py-0.5">
                  {log.length}
                </span>
              )}
            </TabsTrigger>
          </TabsList>

          {/* Events tab */}
          <TabsContent value="events" className="flex-1 overflow-auto px-4 py-3 mt-0">
            <div className="flex flex-col gap-3">
              {socketio.events.map((event) => (
                <div key={event.id} className="border border-border rounded-md p-3 flex flex-col gap-2">
                  <div className="flex items-center gap-2">
                    <Input
                      className="flex-1 h-7 text-xs font-mono"
                      value={event.name}
                      onChange={(e) => updateEvent(event.id, { name: e.target.value })}
                      placeholder="event name"
                    />
                    <Button
                      size="sm"
                      className="h-7 text-xs shrink-0"
                      disabled={status !== 'connected'}
                      onClick={() => emitEvent(event)}
                    >
                      Emit
                    </Button>
                    <Button
                      size="icon"
                      variant="ghost"
                      className="h-7 w-7 shrink-0 text-muted-foreground hover:text-rose-500"
                      onClick={() => deleteEvent(event.id)}
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </div>
                  <Textarea
                    className="text-xs font-mono resize-none"
                    rows={3}
                    value={event.payload}
                    onChange={(e) => updateEvent(event.id, { payload: e.target.value })}
                    placeholder='{"key": "value"}'
                  />
                  <label className="flex items-center gap-2 text-xs text-muted-foreground cursor-pointer select-none">
                    <input
                      type="checkbox"
                      checked={event.useAck}
                      onChange={(e) => updateEvent(event.id, { useAck: e.target.checked })}
                      className="w-3.5 h-3.5"
                    />
                    Expect ack
                  </label>
                </div>
              ))}

              <Button
                size="sm"
                variant="outline"
                className="self-start h-7 text-xs gap-1"
                onClick={addEvent}
              >
                <Plus className="w-3 h-3" />
                Add Event
              </Button>
            </div>
          </TabsContent>

          {/* Log tab */}
          <TabsContent value="log" className="flex-1 overflow-hidden mt-0 flex flex-col">
            <div className="flex justify-end px-4 pt-2 pb-1">
              <Button
                size="sm"
                variant="ghost"
                className="h-6 text-xs"
                onClick={() => setLog([])}
              >
                Clear
              </Button>
            </div>
            <ScrollArea className="flex-1 px-4">
              {log.length === 0 ? (
                <p className="text-xs text-muted-foreground py-4">No events yet.</p>
              ) : (
                <div className="flex flex-col gap-1 pb-4">
                  {log.map((entry) => (
                    <div key={entry.id} className="flex items-start gap-2 text-xs font-mono">
                      <span className="text-muted-foreground shrink-0">{formatTime(entry.ts)}</span>
                      <span className={cn('font-bold shrink-0 w-4', LOG_COLORS[entry.direction])}>
                        {LOG_ARROWS[entry.direction]}
                      </span>
                      <span className={cn('shrink-0 font-semibold', LOG_COLORS[entry.direction])}>
                        {entry.event}
                      </span>
                      <span className="text-muted-foreground truncate">
                        {truncate(entry.payload)}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </ScrollArea>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
