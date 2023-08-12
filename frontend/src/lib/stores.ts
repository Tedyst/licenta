import { persisted } from 'svelte-local-storage-store';
import type { User } from '../interfaces';

export const theme = persisted('theme', 'light');
export const user = persisted<User | null>('user', null);
