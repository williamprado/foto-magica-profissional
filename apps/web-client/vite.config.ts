import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "node:path";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@foto-magica/ui": path.resolve(__dirname, "../../packages/ui/src"),
      "@foto-magica/hooks": path.resolve(__dirname, "../../packages/hooks/src"),
      "@foto-magica/api-client": path.resolve(__dirname, "../../packages/api-client/src"),
      "@foto-magica/types": path.resolve(__dirname, "../../packages/types/src")
    }
  }
});

