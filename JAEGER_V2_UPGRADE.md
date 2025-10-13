# Upgrade do Jaeger para v2

## 📋 Resumo

Atualização do Jaeger de v1 (deprecated) para v2.2.0 para garantir suporte contínuo e acesso a novos recursos.

## 🔄 Mudanças Realizadas

### docker-compose.yml

**Antes:**
```yaml
jaeger:
  image: jaegertracing/all-in-one:latest
  ports:
    - "16686:16686"  # Jaeger UI
    - "14268:14268"  # jaeger.thrift
  environment:
    - COLLECTOR_OTLP_ENABLED=true
```

**Depois:**
```yaml
jaeger:
  image: jaegertracing/jaeger:2.2.0
  ports:
    - "16686:16686"  # Jaeger UI
    - "4317:4317"    # OTLP gRPC receiver
    - "4318:4318"    # OTLP HTTP receiver
    - "14268:14268"  # jaeger.thrift (legacy compatibility)
  environment:
    - COLLECTOR_OTLP_ENABLED=true
    - SPAN_STORAGE_TYPE=memory
```

## 📝 Principais Diferenças

1. **Imagem**: Mudou de `jaegertracing/all-in-one:latest` para `jaegertracing/jaeger:2.2.0`
2. **Portas Adicionais**: 
   - `4317`: OTLP gRPC receiver (protocolo padrão OpenTelemetry)
   - `4318`: OTLP HTTP receiver
3. **Compatibilidade**: Porta 14268 mantida para compatibilidade com aplicações legadas

## 🚀 Endpoints Disponíveis

- **Jaeger UI**: http://localhost:16686
- **OTLP gRPC**: localhost:4317
- **OTLP HTTP**: http://localhost:4318
- **Jaeger Thrift (legacy)**: http://localhost:14268

## ✅ Verificação

Para verificar se o Jaeger v2 está rodando corretamente:

```powershell
# Verificar status do container
docker-compose ps

# Ver logs do Jaeger
docker logs dashtrack-jaeger

# Acessar a UI
start http://localhost:16686
```

## 📚 Referências

- [Jaeger v2 Migration Guide](https://www.jaegertracing.io/docs/latest/migration/)
- [Jaeger v2 GitHub Issue](https://github.com/jaegertracing/jaeger/issues/6321)
- [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)

## ⚠️ Notas Importantes

- Jaeger v1 atinge end-of-life em **31 de dezembro de 2025**
- Jaeger v2 é baseado no OpenTelemetry Collector
- Suporte completo para protocolos OTLP (gRPC e HTTP)
- Melhor integração com ecossistema OpenTelemetry

## 🔧 Comandos Úteis

```powershell
# Parar e remover containers
docker-compose down

# Remover imagem antiga (se necessário)
docker rmi jaegertracing/all-in-one:latest

# Baixar nova imagem
docker-compose pull jaeger

# Subir containers
docker-compose up -d

# Ver logs em tempo real
docker logs dashtrack-jaeger -f
```

---

**Data da Atualização**: 13 de outubro de 2025
**Versão Anterior**: Jaeger v1 (all-in-one:latest)
**Versão Atual**: Jaeger v2.2.0
