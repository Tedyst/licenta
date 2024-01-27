import { persisted } from 'svelte-local-storage-store';
import type { components } from '../lib/api/v1';

export const theme = persisted('theme', 'light');
export const user = persisted<components['schemas']['User'] | null>('user', null);
