SITE="production"
# SITE="certin-site"

echo "GENERATE: $SITE"

# Generate certin content
./clz-md2hugo build \
   -i examples/${SITE} \
   -o examples/generated/content.json
#   -o examples/generated/content.json \
#   --include-unpublished

# Generate catalogue
# Use: `--include-unpublished` includes: Published + draft + review + archived
./clz-md2hugo project \
  --input examples/${SITE} \
  --catalogue examples/generated/catalogue.json \
  --payload-dir examples/generated/private/payloads
#  --payload-dir examples/generated/private/payloads \
#  --include-unpublished

# Copy updated content to destination 
cp examples/generated/content.json ../clz-certin/assets/data
cp examples/generated/catalogue.json ../clz-certin/assets/data
cp examples/generated/private/payloads/public.json ../clz-certin/assets/data
