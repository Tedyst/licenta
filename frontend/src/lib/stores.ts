import { persisted } from 'svelte-local-storage-store';
import type { components } from '../lib/api/v1';
import { writable } from 'svelte/store';

export const theme = persisted('theme', 'light');
export const user = writable<components['schemas']['User'] | null>(null);
export const organizations = writable<components['schemas']['Organization'][] | null>(null);

export const currentOrganization = writable<components['schemas']['Organization'] | null>(null);
export const currentProject = writable<components['schemas']['Project'] | null>(null);

export const currentMysqlDatabases = writable<components['schemas']['MysqlDatabase'][]>([]);
export const currentPostgresDatabases = writable<components['schemas']['PostgresDatabase'][]>([]);
