import { useState, useEffect } from 'react';
import './App.css';
import { Sidebar } from '@/components/sidebar/Sidebar';
import { RequestEditor } from '@/components/request/RequestEditor';
import type { ApiRequest, Folder, RequestType } from '@/types';
import { parsePostman } from '@/lib/parsePostman';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

const STORAGE_KEY = 'api-tester:state';
const BASIC_AUTH_KEY = 'api-tester:basic-auth';

interface PersistedState {
  folders: Folder[];
  requests: ApiRequest[];
  selectedId: string | null;
  envVars: Record<string, string>;
}

interface BasicAuth {
  url: string;
  user: string;
  pass: string;
}

function loadBasicAuth(): BasicAuth {
  try {
    const raw = localStorage.getItem(BASIC_AUTH_KEY);
    if (raw) return JSON.parse(raw) as BasicAuth;
  } catch { /* ignore */ }
  return { url: 'http://localhost:9100', user: '', pass: '' };
}

function makeAuthHeader(auth: BasicAuth): string {
  return 'Basic ' + btoa(`${auth.user}:${auth.pass}`);
}

async function fetchCollection(auth: BasicAuth): Promise<unknown> {
  const res = await fetch(`${auth.url}/collection`, {
    headers: { Authorization: makeAuthHeader(auth) },
  });
  if (!res.ok) throw new Error(`apiwatch responded ${res.status}`);
  const body = await res.json() as { content: unknown };
  return body.content;
}

function loadFromStorage(): PersistedState | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as PersistedState;
  } catch {
    return null;
  }
}

const DEFAULT_BODY = {
  type: 'none' as const,
  json: '',
  multipart: [],
  urlencoded: [],
};

function App() {
  const [folders, setFolders] = useState<Folder[]>([]);
  const [requests, setRequests] = useState<ApiRequest[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [envVars, setEnvVars] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [showReloadConfirm, setShowReloadConfirm] = useState(false);
  const [showBasicAuth, setShowBasicAuth] = useState(false);
  const [basicAuth, setBasicAuth] = useState<BasicAuth>(loadBasicAuth);
  const [basicAuthDraft, setBasicAuthDraft] = useState<BasicAuth>({ url: '', user: '', pass: '' });

  useEffect(() => {
    const saved = loadFromStorage();
    if (saved) {
      setFolders(saved.folders);
      setRequests(saved.requests);
      setSelectedId(saved.selectedId);
      setEnvVars(saved.envVars);
      setLoading(false);
      return;
    }

    const auth = loadBasicAuth();
    fetchCollection(auth)
      .then((json) => {
        const { folders, requests, envVars } = parsePostman(json);
        setFolders(folders);
        setRequests(requests);
        setEnvVars(envVars);
        if (requests.length > 0) setSelectedId(requests[0].id);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (loading) return;
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ folders, requests, selectedId, envVars }));
    } catch {
      // storage quota exceeded or unavailable
    }
  }, [folders, requests, selectedId, envVars, loading]);

  useEffect(() => {
    try {
      localStorage.setItem(BASIC_AUTH_KEY, JSON.stringify(basicAuth));
    } catch { /* ignore */ }
  }, [basicAuth]);

  function openBasicAuthDialog() {
    setBasicAuthDraft({ ...basicAuth });
    setShowBasicAuth(true);
  }

  function saveBasicAuth() {
    setBasicAuth(basicAuthDraft);
    setShowBasicAuth(false);
  }

  const selectedRequest = requests.find((r) => r.id === selectedId) ?? null;

  function toggleFolder(id: string) {
    setFolders((prev) =>
      prev.map((f) => (f.id === id ? { ...f, isOpen: !f.isOpen } : f)),
    );
  }

  function addFolder() {
    const folder: Folder = {
      id: crypto.randomUUID(),
      name: 'New Folder',
      isOpen: true,
    };
    setFolders((prev) => [...prev, folder]);
  }

  function addRequest(folderId: string | null, type: RequestType) {
    const req: ApiRequest = {
      id: crypto.randomUUID(),
      name: type === 'socketio' ? 'New Socket.IO' : type === 'websocket' ? 'New WebSocket' : 'New Request',
      folderId,
      fromPostman: false,
      type,
      method: 'GET',
      url: '',
      headers: [],
      body: { ...DEFAULT_BODY },
      pathVars: {},
      queryParams: [],
      socketio: { url: '', namespace: '/', events: [] },
    };
    setRequests((prev) => [...prev, req]);
    setSelectedId(req.id);
  }

  function updateRequest(updated: ApiRequest) {
    setRequests((prev) => prev.map((r) => (r.id === updated.id ? updated : r)));
  }

  function reloadCollection() {
    localStorage.removeItem(STORAGE_KEY);
    setLoading(true);
    setShowReloadConfirm(false);
    fetchCollection(basicAuth)
      .then((json) => {
        const { folders, requests, envVars } = parsePostman(json);
        setFolders(folders);
        setRequests(requests);
        setEnvVars(envVars);
        setSelectedId(requests.length > 0 ? requests[0].id : null);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar
        folders={folders}
        requests={requests}
        selectedId={selectedId}
        onSelectRequest={setSelectedId}
        onToggleFolder={toggleFolder}
        onAddRequest={addRequest}
        onAddFolder={addFolder}
        onReloadCollection={() => setShowReloadConfirm(true)}
        onOpenBasicAuth={openBasicAuthDialog}
        basicAuthUser={basicAuth.user}
        loading={loading}
      />
      <Dialog open={showReloadConfirm} onOpenChange={setShowReloadConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reload Postman Collection</DialogTitle>
            <DialogDescription>
              This will discard all your current changes — including any edits to URLs, params,
              headers, body, and any requests you added manually — and reload the original
              Postman collection. This cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowReloadConfirm(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={reloadCollection}>
              Reload Collection
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog open={showBasicAuth} onOpenChange={setShowBasicAuth}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Basic Auth Credentials</DialogTitle>
            <DialogDescription>
              These credentials are saved to local storage and used for authenticated requests.
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-3 py-1">
            <div className="flex flex-col gap-1">
              <label className="text-xs font-medium text-muted-foreground">Server URL</label>
              <Input
                value={basicAuthDraft.url}
                onChange={(e) => setBasicAuthDraft((d) => ({ ...d, url: e.target.value }))}
                placeholder="http://localhost:9100"
                autoComplete="off"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs font-medium text-muted-foreground">Username</label>
              <Input
                value={basicAuthDraft.user}
                onChange={(e) => setBasicAuthDraft((d) => ({ ...d, user: e.target.value }))}
                placeholder="username"
                autoComplete="username"
              />
            </div>
            <div className="flex flex-col gap-1">
              <label className="text-xs font-medium text-muted-foreground">Password</label>
              <Input
                type="password"
                value={basicAuthDraft.pass}
                onChange={(e) => setBasicAuthDraft((d) => ({ ...d, pass: e.target.value }))}
                placeholder="password"
                autoComplete="current-password"
                onKeyDown={(e) => e.key === 'Enter' && saveBasicAuth()}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowBasicAuth(false)}>Cancel</Button>
            <Button onClick={saveBasicAuth}>Save</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <main className="flex-1 overflow-hidden">
        {loading ? (
          <div className="flex h-full items-center justify-center text-muted-foreground text-sm">
            Loading collection…
          </div>
        ) : selectedRequest ? (
          <RequestEditor request={selectedRequest} onChange={updateRequest} envVars={envVars} />
        ) : (
          <div className="flex h-full items-center justify-center text-muted-foreground text-sm">
            Select a request from the sidebar
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
