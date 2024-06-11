import type { Actions } from './$types';
import { clientFromFetch } from '$lib/client';
import { redirect } from '@sveltejs/kit';

export const actions = {
	default: async ({ request, fetch, url }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch, url.origin);

		const projectId = data.get('projectId')?.toString();
		const hostname = data.get('hostname')?.toString();
		const port = data.get('port')?.toString();
		const username = data.get('username')?.toString();
		const password = data.get('password')?.toString();

		const organizationName = data.get('organizationName')?.toString();
		const projectName = data.get('projectName')?.toString();

		if (projectId === '0' || projectId === undefined) {
			return { error: 'Project is required' };
		}
		if (hostname === '' || hostname === undefined) {
			return { error: 'Hostname is required' };
		}
		if (port === '0' || port === undefined) {
			return { error: 'Port is required' };
		}
		if (username === '' || username === undefined) {
			return { error: 'User is required' };
		}
		if (password === '' || password === undefined) {
			return { error: 'Password is required' };
		}
		if (organizationName === '' || organizationName === undefined) {
			return { error: 'Organization name is required' };
		}
		if (projectName === '' || projectName === undefined) {
			return { error: 'Project name is required' };
		}

		const returned = await client
			.POST('/redis', {
				body: {
					host: hostname,
					password: password,
					port: +port,
					project_id: +projectId,
					username: username
				}
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
