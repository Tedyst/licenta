import { persisted } from 'svelte-local-storage-store';
import type { components } from '../lib/api/v1';
import { writable } from 'svelte/store';

export const theme = persisted('theme', 'light');
export const user = writable<components['schemas']['User'] | null>(null);
export const organizations = writable<components['schemas']['Organization'][]>([]);

export const currentOrganization = writable<components['schemas']['Organization'] | null>(null);

// setTimeout(
// 	() =>
// 		organizations.set([
// 			{
// 				id: 1,
// 				name: 'test-org',
// 				created_at: '',
// 				projects: [
// 					{ id: 1, name: 'test-prog1111-11111', created_at: '', organization_id: 1, remote: false }
// 				]
// 			},
// 			{
// 				id: 2,
// 				name: 'test-org-222-222',
// 				created_at: '',
// 				projects: [{ id: 2, name: 'test-prog2', created_at: '', organization_id: 1, remote: false }]
// 			}
// 		]),
// 	10000
// );
