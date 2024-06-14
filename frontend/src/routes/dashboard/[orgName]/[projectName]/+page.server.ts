import type { Actions } from './$types';
import { clientFromFetch } from '$lib/client';
import { invalidateAll } from '$app/navigation';
import { env } from '$env/dynamic/public';
invalidateAll

export const actions = {
	run: async ({ request, fetch, url }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch, env.PUBLIC_BACKEND_URL);

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
	}
} satisfies Actions;
