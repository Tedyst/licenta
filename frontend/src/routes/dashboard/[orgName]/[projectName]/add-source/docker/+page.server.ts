import type { Actions } from './$types';
import { clientFromFetch } from '$lib/client';
import { redirect } from '@sveltejs/kit';

export const actions = {
	default: async ({ request, fetch, url }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch, url.origin);

		const image = data.get('image')?.toString();
		const username = data.get('username')?.toString();
		const password = data.get('password')?.toString();

		const projectId = data.get('projectId')?.toString();
		const organizationName = data.get('organizationName')?.toString();
		const projectName = data.get('projectName')?.toString();

		if (projectId === '0' || projectId === undefined) {
			return { error: 'Project is required' };
		}
		if (image === '' || image === undefined) {
			return { error: 'Image is required' };
		}
		if (organizationName === '' || organizationName === undefined) {
			return { error: 'Organization name is required' };
		}
		if (projectName === '' || projectName === undefined) {
			return { error: 'Project name is required' };
		}

		const returned = await client
			.POST('/docker', {
				body: {
					docker_image: image,
					password: password === '' ? undefined : password,
					project_id: +projectId,
					username: username === '' ? undefined : username
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
