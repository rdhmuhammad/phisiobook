import { useRef, useState, useEffect } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Textarea } from '@/components/ui/textarea';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';
import type { ApiRequest, RequestType } from '@/types';

type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error';
type LogDirection = 'sent' | 'received' | 'error';

interface LogEntry {
  id: string;
  ts: Date;
  direction: LogDirection;
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
  error: 'text-rose-600',
};

const LOG_ARROWS: Record<LogDirection, string> = {
  sent: '→',
  received: '←',
  error: '✗',
};

const TYPE_OPTIONS: { value: RequestType; label: string }[] = [
  { value: 'rest', label: 'REST' },
  { value: 'socketio', label: 'Socket.IO' },
  { value: 'websocket', label: 'WebSocket' },
];

interface WebSocketEditorProps {
  request: ApiRequest;
  onChange: (updated: ApiRequest) => void;
}

export function WebSocketEditor({ request, onChange }: WebSocketEditorProps) {
  const wsRef = useRef<WebSocket | null>(null);
  const [status, setStatus] = useState<ConnectionStatus>('disconnected');
  const [log, setLog] = useState<LogEntry[]>([]);
  const [message, setMessage] = useState('');

  useEffect(() => {
    return () => {
      wsRef.current?.close();
    };
  }, []);

  function addLog(direction: LogDirection, payload: string) {
    setLog((prev) => [
      { id: crypto.randomUUID(), ts: new Date(), direction, payload },
      ...prev,
    ]);
  }

  function connect() {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setStatus('connecting');
    try {
      const ws = new WebSocket(request.socketio.url);
      wsRef.current = ws;
      ws.onopen = () => setStatus('connected');
      ws.onmessage = (ev) => addLog('received', typeof ev.data === 'string' ? ev.data : JSON.stringify(ev.data));
      ws.onerror = () => { setStatus('error'); addLog('error', 'Connection error'); };
      ws.onclose = () => { setStatus('disconnected'); wsRef.current = null; };
    } catch (err) {
      setStatus('error');
      addLog('error', String(err));
    }
  }

  function disconnect() {
    wsRef.current?.close();
    wsRef.current = null;
    setStatus('disconnected');
  }

  function sendMessage() {
    if (!wsRef.current || status !== 'connected' || !message.trim()) return;
    wsRef.current.send(message);
    addLog('sent', message);
    setMessage('');
  }

  function formatTime(ts: Date) {
    return ts.toTimeString().slice(0, 8);
  }

  function truncate(str: string, max = 120) {
    return str.length > max ? str.slice(0, max) + '…' : str;
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

      {/* URL + connect */}
      <div className="flex items-center gap-2 px-4 py-3">
        <Input
          className="flex-1 h-8 text-sm font-mono"
          value={request.socketio.url}
          onChange={(e) => onChange({ ...request, socketio: { ...request.socketio, url: e.target.value } })}
          placeholder="ws://localhost:3001"
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

      {/* Message input */}
      <div className="flex gap-2 px-4 py-3">
        <Textarea
          className="flex-1 text-xs font-mono resize-none"
          rows={2}
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="Message to send…"
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault();
              sendMessage();
            }
          }}
        />
        <Button
          size="sm"
          className="h-auto self-stretch"
          disabled={status !== 'connected' || !message.trim()}
          onClick={sendMessage}
        >
          Send
        </Button>
      </div>

      <Separator />

      {/* Log */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="flex justify-end px-4 pt-2 pb-1 shrink-0">
          <Button size="sm" variant="ghost" className="h-6 text-xs" onClick={() => setLog([])}>
            Clear
          </Button>
        </div>
        <ScrollArea className="flex-1 px-4">
          {log.length === 0 ? (
            <p className="text-xs text-muted-foreground py-4">No messages yet.</p>
          ) : (
            <div className="flex flex-col gap-1 pb-4">
              {log.map((entry) => (
                <div key={entry.id} className="flex items-start gap-2 text-xs font-mono">
                  <span className="text-muted-foreground shrink-0">{formatTime(entry.ts)}</span>
                  <span className={cn('font-bold shrink-0 w-4', LOG_COLORS[entry.direction])}>
                    {LOG_ARROWS[entry.direction]}
                  </span>
                  <span className={cn('text-muted-foreground truncate', LOG_COLORS[entry.direction])}>
                    {truncate(entry.payload)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
      </div>
    </div>
  );
}
