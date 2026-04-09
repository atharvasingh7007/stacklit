import express from 'express';
import { Router } from './router';
import type { Config } from './types';

const app = express();

export function startServer(config: Config) {
  app.use(Router);
  app.listen(config.port);
}

export default app;
