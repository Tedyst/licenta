import type { Actions } from './$types';
import { clientFromFetch } from '$lib/client';
import { redirect } from '@sveltejs/kit';

export const actions = {
	default: async ({ request, fetch, url }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch, url.origin);

		const organizationName = data.get('organizationName')?.valueOf();

		if (typeof organizationName !== 'string' || organizationName.trim() === '') {
			return { error: 'Organization name is required' };
		}

		const returned = await client
			.POST('/organizations', { body: { name: organizationName.trim().toLowerCase() } })
			.then(async (res) => {
				if (res.data?.success) {
					return {
						url: `/dashboard/${res.data.organization.name}`,
						error: null
					};
				}
				return { error: res.error?.message || 'Internal server error' };
			})
			.catch((err) => {
				return { error: err.message || 'Internal server error' };
			});

		if (returned.error === null) {
			redirect(302, returned.url);
		}

		return returned;
	}
} satisfies Actions;
