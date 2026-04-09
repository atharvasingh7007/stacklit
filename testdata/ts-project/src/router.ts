import { Router as ExpressRouter } from 'express';
const router = require('./middleware');

export class AppRouter {
  setup() {}
}

export const Router = ExpressRouter();
export type RouteHandler = (req: any, res: any) => void;
