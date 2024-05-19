import type { Actions } from './$types';
import { clientFromFetch } from '$lib/client';
import { redirect } from '@sveltejs/kit';

export const actions = {
	editRole: async ({ request, fetch, url }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch, url.origin);

		const organizationId = data.get('organizationId')?.toString();
		const userId = data.get('userId')?.toString();
		const role = data.get('role')?.toString();

		if (organizationId === '0' || organizationId === undefined) {
			return { error: 'Organization name is required' };
		}
		if (userId === '0' || userId === undefined) {
			return { error: 'User is required' };
		}
		if (role === '0' || role === undefined) {
			return { error: 'Role is required' };
		}

		const returned = await client
			.POST('/organizations/{id}/edit-user', {
				params: { path: { id: +organizationId } },
				body: { id: +userId, role }
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

		redirect(302, '/dashboard/' + organizationId);
	},
	delete: async ({ request, fetch, url }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch, url.origin);

		const organizationId = data.get('organizationId')?.toString();
		const userId = data.get('userId')?.toString();

		if (organizationId === '0' || organizationId === undefined) {
			return { error: 'Organization name is required' };
		}
		if (userId === '0' || userId === undefined) {
			return { error: 'User is required' };
		}

		const returned = await client
			.DELETE('/organizations/{id}/delete-user', {
				params: { path: { id: +organizationId } },
				body: { id: +userId }
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

		redirect(302, '/dashboard/' + organizationId);
	}
} satisfies Actions;
