import type { Actions } from './$types';
import { clientFromFetch } from '$lib/client';
import { redirect } from '@sveltejs/kit';

export const actions = {
	default: async ({ request, fetch, url }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch, url.origin);

		const organizationId = data.get('organizationId')?.toString();

		if (organizationId === '0' || organizationId === undefined) {
			return { error: 'Organization name is required' };
		}

		const returned = await client
			.DELETE('/organizations/{id}', {
				params: { path: { id: +organizationId } }
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

		redirect(302, '/dashboard');
	}
} satisfies Actions;
