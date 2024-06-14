import type { Actions } from './$types';
import { clientFromFetch } from '$lib/client';
import { redirect } from '@sveltejs/kit';

export const actions = {
	default: async ({ request, fetch, url }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch);

		const organizationId = data.get('organizationId')?.toString();
		const organizationName = data.get('organizationName')?.toString();
		const projectName = data.get('projectName')?.toString();

		if (organizationId === '0' || organizationId === undefined) {
			return { error: 'Organization name is required' };
		}
		if (projectName === '' || projectName === undefined) {
			return { error: 'Project name is required' };
		}
		if (organizationName === '' || organizationName === undefined) {
			return { error: 'Organization name is required' };
		}

		const returned = await client
			.POST('/projects', {
				body: { name: projectName.toLowerCase(), organization_id: +organizationId }
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

		redirect(302, '/dashboard/' + organizationName + '/' + projectName);
	}
} satisfies Actions;
