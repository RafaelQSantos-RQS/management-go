# Docker Cleanup Script

Script automatizado em Go para limpeza de ambientes Docker.

## O que faz

Remove automaticamente:
- ✅ Containers parados
- ✅ Volumes não utilizados
- ✅ Redes não utilizadas
- ✅ Imagens não utilizadas (incluindo com tags)

**Segurança:** Containers em execução e suas imagens são sempre preservados.

## Como usar

### Opção 1: Executar com Go
```bash
go run main.go
```

### Opção 2: Compilar e executar binário
```bash
# Compilar
go build -o docker-cleanup main.go

# Executar
./docker-cleanup
```

### Opção 4: Executar com Docker (Recomendado para portabilidade)
```bash
# Build da imagem
docker build -t docker-cleanup .

# Executar como Daemon (Recorrente)
Para que a limpeza ocorra automaticamente sem cron externo, use a variável `CLEANUP_INTERVAL`.

```bash
docker run -d \
  --name docker-cleanup \
  -e CLEANUP_INTERVAL=24h \
  -v /var/run/docker.sock:/var/run/docker.sock \
  docker-cleanup
```

Valores válidos: `1h`, `24h`, `30m`, `1d`. Se não for definida, o script roda uma vez e sai.

### Opção 5: Adicionar ao cron para manutenção automática
```bash
# Editar crontab
crontab -e

# Executar toda segunda-feira às 3h da manhã
0 3 * * 1 /caminho/para/docker-cleanup >> /var/log/docker-cleanup.log 2>&1

# Ou executar diariamente às 2h da manhã
0 2 * * * /caminho/para/docker-cleanup >> /var/log/docker-cleanup.log 2>&1
```

## Exemplo de saída

```
🧹 Docker Cleanup Script - Iniciando limpeza automática...
========================================================

🗑️  Removendo containers parados...
   ℹ️  Nenhum container parado encontrado

🗑️  Removendo volumes não utilizados...
   ℹ️  Nenhum volume não utilizado encontrado

🗑️  Removendo redes não utilizadas...
   ℹ️  Nenhuma rede não utilizada encontrada

🗑️  Removendo imagens não utilizadas...
   Removendo: postgres:latest (ID: sha256:38d5c, Tamanho: 618.97 MB)
   Removendo: timescale/timescaledb:latest-pg17 (ID: sha256:e9532, Tamanho: 1139.59 MB)
   ✅ 2 imagens removidas
   💾 Espaço recuperado: 1758.56 MB

========================================================
✅ Limpeza completa finalizada!
```

## Requisitos

- Go 1.25+
- Acesso ao Docker socket (geralmente `/var/run/docker.sock`)
- Permissões para executar comandos Docker

## Instalação em servidor

```bash
# Clone ou copie o projeto
cd /opt
git clone <seu-repo> docker-cleanup
cd docker-cleanup

# Compile
go build -o docker-cleanup main.go

# Torne executável
chmod +x docker-cleanup

# (Opcional) Crie link simbólico para usar globalmente
sudo ln -s /opt/docker-cleanup/docker-cleanup /usr/local/bin/docker-cleanup
```
