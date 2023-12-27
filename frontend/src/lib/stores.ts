import { persisted } from 'svelte-local-storage-store';
import type { User } from '../interfaces';
import { writable } from 'svelte/store';

export const theme = persisted('theme', 'light');
export const user = writable<User | null>(null);
