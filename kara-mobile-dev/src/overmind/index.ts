// overmind/index.ts
import {IContext, createOvermind} from 'overmind';
import {
  createStateHook,
  createActionsHook,
  createEffectsHook,
  createReactionHook,
} from 'overmind-react';
import {state} from './state';
import * as actions from './actions';
import * as effects from './effects';
import {OVERMIND_HOST} from '@env';

export const config = {
  state,
  actions,
  effects,
};

export type Context = IContext<typeof config>;

export const useAppState = createStateHook<Context>();
export const useActions = createActionsHook<Context>();
export const useEffects = createEffectsHook<Context>();
export const useReaction = createReactionHook<Context>();
export const store = createOvermind(config, {
  devtools: __DEV__ ? `${OVERMIND_HOST}:3031` : false,
});
