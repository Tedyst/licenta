import { redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { clientFromFetch } from '$lib/client';

export const load: PageServerLoad = async ({ parent, url }) => {
	const parentData = await parent();

	const databaseId = url.searchParams.get('id') || '0';

	if (databaseId === '0') {
		redirect(302, '/dashboard/' + parentData.organization?.name + '/' + parentData.project?.name);
	}

	const currentDatabase =
		parentData.redisDatabases?.filter((v) => v.id === +databaseId).at(0) || null;

	if (!currentDatabase) {
		redirect(302, '/dashboard/' + parentData.organization?.name + '/' + parentData.project?.name);
	}

	return {
		...parentData,
		currentDatabase: currentDatabase
	};
};

export const actions = {
	default: async ({ request, fetch, url }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch, url.origin);

		const databaseId = data.get('databaseId')?.toString();
		const hostname = data.get('hostname')?.toString();
		const port = data.get('port')?.toString();
		const username = data.get('username')?.toString();
		const password = data.get('password')?.toString();

		const organizationName = data.get('organizationName')?.toString();
		const projectName = data.get('projectName')?.toString();

		if (databaseId === '0' || databaseId === undefined) {
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
		if (organizationName === '' || organizationName === undefined) {
			return { error: 'Organization name is required' };
		}
		if (projectName === '' || projectName === undefined) {
			return { error: 'Project name is required' };
		}

		const returned = await client
			.PATCH('/redis/{id}', {
				body: {
					host: hostname,
					password: password === 'pass' ? undefined : password,
					port: +port,
					username: username
				},
				params: { path: { id: +databaseId } }
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
