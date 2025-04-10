// src/env.d.ts
interface ImportMetaEnv {
    readonly VITE_API_BASE_URL: string;
    // other env variables...
}

interface ImportMeta {
    readonly env: ImportMetaEnv;
}
