import type { Actions } from './$types';
import { clientFromFetch } from '$lib/client';
import { redirect } from '@sveltejs/kit';

export const actions = {
	default: async ({ request, fetch, url }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch, url.origin);

		const projectId = data.get('projectId')?.toString();
		const organizationName = data.get('organizationName')?.toString();

		if (projectId === '0' || projectId === undefined) {
			return { error: 'Project name is required' };
		}
		if (organizationName === '' || organizationName === undefined) {
			return { error: 'Organization name is required' };
		}

		const returned = await client
			.DELETE('/projects/{id}', {
				params: { path: { id: +projectId } }
			})
			.then((res) => {
				if (res.response.ok) {
					return {};
				}
				return { error: res.error?.message || 'An error occurred' };
			});

		if (returned.error) {
			return { error: returned.error };
		}

		redirect(302, '/dashboard/' + organizationName);
	}
} satisfies Actions;
