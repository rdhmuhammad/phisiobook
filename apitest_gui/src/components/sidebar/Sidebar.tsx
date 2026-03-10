import { Folder as FolderIcon, FolderPlus, Plus, Zap, ChevronRight, ChevronDown, RefreshCw, KeyRound } from 'lucide-react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import type { ApiRequest, Folder, HttpMethod, RequestType } from '@/types';

const METHOD_COLORS: Record<HttpMethod, string> = {
  GET: 'text-emerald-600',
  POST: 'text-sky-600',
  PUT: 'text-amber-600',
  PATCH: 'text-violet-600',
  DELETE: 'text-rose-600',
  HEAD: 'text-slate-500',
  OPTIONS: 'text-slate-500',
};

function RequestBadge({ request }: { request: ApiRequest }) {
  if (request.type === 'socketio') {
    return <span className="text-xs font-bold w-[42px] shrink-0 text-violet-600">IO</span>;
  }
  if (request.type === 'websocket') {
    return <span className="text-xs font-bold w-[42px] shrink-0 text-amber-600">WS</span>;
  }
  return (
    <span className={cn('text-xs font-bold w-[42px] shrink-0', METHOD_COLORS[request.method])}>
      {request.method}
    </span>
  );
}

interface SidebarProps {
  folders: Folder[];
  requests: ApiRequest[];
  selectedId: string | null;
  onSelectRequest: (id: string) => void;
  onToggleFolder: (id: string) => void;
  onAddRequest: (folderId: string | null, type: RequestType) => void;
  onAddFolder: () => void;
  onReloadCollection: () => void;
  onOpenBasicAuth: () => void;
  basicAuthUser?: string;
  loading?: boolean;
}

export function Sidebar({
  folders,
  requests,
  selectedId,
  onSelectRequest,
  onToggleFolder,
  onAddRequest,
  onAddFolder,
  onReloadCollection,
  onOpenBasicAuth,
  basicAuthUser,
  loading,
}: SidebarProps) {
  const topLevel = requests.filter((r) => r.folderId === null);

  return (
    <aside className="flex flex-col w-[260px] shrink-0 border-r border-border bg-muted/30 h-full">
      {/* Header */}
      <div className="flex items-center gap-2 px-3 py-3 border-b border-border">
        <Zap className="w-4 h-4 text-amber-500" />
        <span className="font-semibold text-sm flex-1">API Tester</span>
        <Button
          size="icon"
          variant="ghost"
          className="h-6 w-6 relative"
          onClick={onOpenBasicAuth}
          title={basicAuthUser ? `Basic auth: ${basicAuthUser}` : 'Set basic auth credentials'}
        >
          <KeyRound className={cn('w-3.5 h-3.5', basicAuthUser ? 'text-amber-500' : 'text-muted-foreground')} />
          {basicAuthUser && (
            <span className="absolute -top-0.5 -right-0.5 w-1.5 h-1.5 rounded-full bg-amber-500" />
          )}
        </Button>
        <Button
          size="icon"
          variant="ghost"
          className="h-6 w-6"
          onClick={onReloadCollection}
          title="Reload Postman collection"
        >
          <RefreshCw className="w-3.5 h-3.5 text-muted-foreground" />
        </Button>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-1 px-2 py-2">
        <Button
          size="sm"
          variant="outline"
          className="flex-1 h-7 text-xs gap-1"
          onClick={() => onAddRequest(null, 'rest')}
        >
          <Plus className="w-3 h-3" />
          REST
        </Button>
        <Button
          size="sm"
          variant="outline"
          className="flex-1 h-7 text-xs gap-1 text-violet-600 border-violet-200 hover:text-violet-700"
          onClick={() => onAddRequest(null, 'socketio')}
        >
          <Plus className="w-3 h-3" />
          WS
        </Button>
        <Button
          size="icon"
          variant="ghost"
          className="h-7 w-7"
          onClick={onAddFolder}
          title="New folder"
        >
          <FolderPlus className="w-4 h-4" />
        </Button>
      </div>

      <Separator />

      {/* List */}
      <ScrollArea className="flex-1">
        <div className="py-1">
        {loading && (
          <p className="text-xs text-muted-foreground px-3 py-2">Loading…</p>
        )}
          {folders.map((folder) => {
            const folderRequests = requests.filter((r) => r.folderId === folder.id);
            return (
              <div key={folder.id}>
                {/* Folder row */}
                <button
                  className="flex items-center w-full gap-1.5 px-2 py-1.5 text-sm hover:bg-accent rounded-sm mx-1 text-left"
                  style={{ width: 'calc(100% - 8px)' }}
                  onClick={() => onToggleFolder(folder.id)}
                >
                  {folder.isOpen ? (
                    <ChevronDown className="w-3.5 h-3.5 shrink-0 text-muted-foreground" />
                  ) : (
                    <ChevronRight className="w-3.5 h-3.5 shrink-0 text-muted-foreground" />
                  )}
                  <FolderIcon className="w-3.5 h-3.5 shrink-0 text-amber-500" />
                  <span className="truncate font-medium">{folder.name}</span>
                  <Button
                    size="icon"
                    variant="ghost"
                    className="ml-auto h-5 w-5 opacity-0 group-hover:opacity-100 hover:opacity-100 shrink-0"
                    onClick={(e) => {
                      e.stopPropagation();
                      onAddRequest(folder.id, 'rest');
                    }}
                    title="Add request to folder"
                  >
                    <Plus className="w-3 h-3" />
                  </Button>
                </button>

                {/* Folder contents */}
                {folder.isOpen &&
                  folderRequests.map((req) => (
                    <RequestRow
                      key={req.id}
                      request={req}
                      selected={selectedId === req.id}
                      onSelect={onSelectRequest}
                      indent
                    />
                  ))}
              </div>
            );
          })}

          {/* Separator between folders and top-level */}
          {folders.length > 0 && topLevel.length > 0 && (
            <Separator className="my-1 mx-2" style={{ width: 'calc(100% - 16px)' }} />
          )}

          {/* Top-level requests */}
          {topLevel.map((req) => (
            <RequestRow
              key={req.id}
              request={req}
              selected={selectedId === req.id}
              onSelect={onSelectRequest}
            />
          ))}
        </div>
      </ScrollArea>
    </aside>
  );
}

function RequestRow({
  request,
  selected,
  onSelect,
  indent = false,
}: {
  request: ApiRequest;
  selected: boolean;
  onSelect: (id: string) => void;
  indent?: boolean;
}) {
  return (
    <button
      className={cn(
        'flex items-center w-full gap-1.5 py-1.5 text-sm rounded-sm mx-1 text-left',
        indent ? 'pl-7 pr-2' : 'px-2',
        selected ? 'bg-accent text-accent-foreground' : 'hover:bg-accent/60',
      )}
      style={{ width: 'calc(100% - 8px)' }}
      onClick={() => onSelect(request.id)}
    >
      <RequestBadge request={request} />
      <span className="truncate">{request.name}</span>
    </button>
  );
}
