all: backend frontend

backend:
	(cd server && go build -o ../dist/server.exe .)

frontend:
	(cd frontend && npm run build && xcopy public "../dist/public" /Y /S)