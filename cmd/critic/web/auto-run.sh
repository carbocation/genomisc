CompileDaemon -directory=./ -command='./carbocation-tools.linux -db_port=5433' -build='go build -o carbocation-tools.linux' -pattern='(.+\.go|.+\.c|.+\.html)$' -graceful-kill=true
