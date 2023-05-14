/// <reference types="vite/client" />

interface ImportMetaEnv {
	readonly VITE_WEBSOCKET_URL: string;
	readonly VITE_WEB_URL: string;
	readonly VITE_LOG_LEVEL: 0;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}
