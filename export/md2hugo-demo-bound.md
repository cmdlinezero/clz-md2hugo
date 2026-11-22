# TermCast MCP: MD2HUGO

{{< step label="Run md2hugo" >}}

### 1. Generate the content

```bash {terminal="true" file-id="md2hugo-demo.jsonl" command-id="6561a78f2e0d180d"}
./clz-md2hugo build \
   -i examples/production \
   -o examples/generated/content.json
```

### 2. Refresh the catalogue

```bash {terminal="true" file-id="md2hugo-demo.jsonl" command-id="11c17273caed6eda"}
./clz-md2hugo project \
  --input examples/production \
  --catalogue examples/generated/catalogue.json \
  --payload-dir examples/generated/private/payloads
```

{{< /step >}}
