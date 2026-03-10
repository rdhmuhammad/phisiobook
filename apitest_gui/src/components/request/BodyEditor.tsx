import { Trash2, Plus } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import type { RequestBody, FormField, BodyType } from '@/types';

interface BodyEditorProps {
  body: RequestBody;
  onChange: (body: RequestBody) => void;
}

export function BodyEditor({ body, onChange }: BodyEditorProps) {
  function setType(type: BodyType) {
    onChange({ ...body, type });
  }

  function formatJson() {
    try {
      const parsed = JSON.parse(body.json);
      onChange({ ...body, json: JSON.stringify(parsed, null, 2) });
    } catch {
      // invalid json — leave as-is
    }
  }

  function updateField(
    field: 'multipart' | 'urlencoded',
    id: string,
    patch: Partial<FormField>,
  ) {
    onChange({
      ...body,
      [field]: body[field].map((f) => (f.id === id ? { ...f, ...patch } : f)),
    });
  }

  function addField(field: 'multipart' | 'urlencoded') {
    const newField: FormField = {
      id: crypto.randomUUID(),
      key: '',
      value: '',
      enabled: true,
    };
    onChange({ ...body, [field]: [...body[field], newField] });
  }

  function removeField(field: 'multipart' | 'urlencoded', id: string) {
    onChange({ ...body, [field]: body[field].filter((f) => f.id !== id) });
  }

  return (
    <Tabs
      value={body.type}
      onValueChange={(v) => setType(v as BodyType)}
      className="flex flex-col h-full"
    >
      <TabsList className="shrink-0 h-8 bg-transparent border-b border-border rounded-none justify-start gap-0 px-0">
        {(['none', 'json', 'multipart', 'urlencoded'] as BodyType[]).map((t) => (
          <TabsTrigger
            key={t}
            value={t}
            className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:shadow-none h-8 px-3 text-xs capitalize"
          >
            {t === 'urlencoded' ? 'URL Encoded' : t === 'none' ? 'None' : t === 'json' ? 'JSON' : 'Multipart'}
          </TabsTrigger>
        ))}
      </TabsList>

      <TabsContent value="none" className="flex-1 flex items-center justify-center mt-0">
        <p className="text-muted-foreground text-sm">This request has no body</p>
      </TabsContent>

      <TabsContent value="json" className="flex-1 flex flex-col gap-2 mt-0 p-2">
        <div className="flex justify-end">
          <Button size="sm" variant="outline" className="h-6 text-xs" onClick={formatJson}>
            Format JSON
          </Button>
        </div>
        <Textarea
          className="flex-1 font-mono text-xs resize-none"
          value={body.json}
          onChange={(e) => onChange({ ...body, json: e.target.value })}
          placeholder='{ "key": "value" }'
          spellCheck={false}
        />
      </TabsContent>

      <TabsContent value="multipart" className="flex-1 flex flex-col mt-0 p-2 overflow-auto">
        <FieldTable
          fields={body.multipart}
          onChange={(id, patch) => updateField('multipart', id, patch)}
          onRemove={(id) => removeField('multipart', id)}
          onAdd={() => addField('multipart')}
        />
      </TabsContent>

      <TabsContent value="urlencoded" className="flex-1 flex flex-col mt-0 p-2 overflow-auto">
        <FieldTable
          fields={body.urlencoded}
          onChange={(id, patch) => updateField('urlencoded', id, patch)}
          onRemove={(id) => removeField('urlencoded', id)}
          onAdd={() => addField('urlencoded')}
        />
      </TabsContent>
    </Tabs>
  );
}

function FieldTable({
  fields,
  onChange,
  onRemove,
  onAdd,
}: {
  fields: FormField[];
  onChange: (id: string, patch: Partial<FormField>) => void;
  onRemove: (id: string) => void;
  onAdd: () => void;
}) {
  return (
    <div className="flex flex-col gap-1">
      {/* Header */}
      <div className="grid grid-cols-[24px_1fr_1fr_28px] gap-1 px-1 text-xs text-muted-foreground font-medium">
        <span />
        <span>Key</span>
        <span>Value</span>
        <span />
      </div>

      {fields.map((field) => (
        <div key={field.id} className="grid grid-cols-[24px_1fr_1fr_28px] gap-1 items-center">
          <input
            type="checkbox"
            checked={field.enabled}
            onChange={(e) => onChange(field.id, { enabled: e.target.checked })}
            className="w-4 h-4 accent-primary"
          />
          <Input
            className="h-7 text-xs"
            value={field.key}
            onChange={(e) => onChange(field.id, { key: e.target.value })}
            placeholder="key"
          />
          <Input
            className="h-7 text-xs"
            value={field.value}
            onChange={(e) => onChange(field.id, { value: e.target.value })}
            placeholder="value"
          />
          <Button
            size="icon"
            variant="ghost"
            className="h-7 w-7"
            onClick={() => onRemove(field.id)}
          >
            <Trash2 className="w-3.5 h-3.5 text-muted-foreground" />
          </Button>
        </div>
      ))}

      <Button
        size="sm"
        variant="ghost"
        className="mt-1 h-7 text-xs self-start gap-1"
        onClick={onAdd}
      >
        <Plus className="w-3 h-3" />
        Add Field
      </Button>
    </div>
  );
}
