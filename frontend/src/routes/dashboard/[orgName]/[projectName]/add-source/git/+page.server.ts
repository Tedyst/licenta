import type { Actions } from './$types';
import { clientFromFetch } from '$lib/client';
import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';

export const actions = {
	default: async ({ request, fetch, url }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch, env.PUBLIC_BACKEND_URL);

		const projectId = data.get('projectId')?.toString();
		const repository = data.get('repository')?.toString();
		const username = data.get('username')?.toString();
		const password = data.get('password')?.toString();
		const privateKey = data.get('privateKey')?.toString();

		const organizationName = data.get('organizationName')?.toString();
		const projectName = data.get('projectName')?.toString();

		if (!projectId) {
			return { error: 'Invalid project ID' };
		}
		if (repository === '' || repository === undefined) {
			return { error: 'Git repository is required' };
		}

		const returned = await client
			.POST('/git', {
				body: {
					git_repository: repository,
					password: password === '' ? undefined : password,
					private_key: privateKey === '' ? undefined : privateKey,
					username: username === '' ? undefined : username,
					project_id: +projectId
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
