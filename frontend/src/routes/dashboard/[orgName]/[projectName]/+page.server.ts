import type { Actions } from './$types';
import { clientFromFetch } from '$lib/client';

export const actions = {
	run: async ({ request, fetch }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch);

		const projectId = data.get('projectId')?.toString();

		if (!projectId) {
			return { error: 'Invalid project ID' };
		}

		const returned = await client
			.POST('/projects/{id}/run', {
				params: { path: { id: +projectId } }
			})
			.then((res) => {
				if (res.data?.success) {
					return {};
				}
				return { error: res.error?.message || 'An error occurred' };
			});

		if (returned.error) {
			return { error: returned.error };
		}

		return {};
	},
	toggle_remote: async ({ request, fetch }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch);

		const projectId = data.get('projectId')?.toString();
		const remoteStr = data.get('remote')?.toString();

		if (!projectId) {
			return { error: 'Invalid project ID' };
		}
		if (!remoteStr) {
			return { error: 'Invalid remote' };
		}

		const returned = await client
			.PATCH('/projects/{id}', {
				params: { path: { id: +projectId } },
				body: { remote: remoteStr === 'true' }
			})
			.then((res) => {
				if (res.data?.success) {
					return {};
				}
				return { error: res.error?.message || 'An error occurred' };
			});

		if (returned.error) {
			return { error: returned.error };
		}

		return {};
	}
} satisfies Actions;
