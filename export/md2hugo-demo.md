# TermCast MCP: MD2HUGO

{{< step label="Run md2hugo" >}}

### 1. Generate the content

```bash {terminal="true"}
./clz-md2hugo build \
   -i examples/production \
   -o examples/generated/content.json
```

### 2. Refresh the catalogue

```bash {terminal="true"}
./clz-md2hugo project \
  --input examples/production \
  --catalogue examples/generated/catalogue.json \
  --payload-dir examples/generated/private/payloads
```

{{< /step >}}
