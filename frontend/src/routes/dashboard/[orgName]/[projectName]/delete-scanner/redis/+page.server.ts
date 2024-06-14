import type { Actions, PageServerLoad } from './$types';
import { clientFromFetch } from '$lib/client';
import { redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';

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
		const client = clientFromFetch(fetch, env.PUBLIC_BACKEND_URL);

		const databaseId = data.get('databaseId')?.toString();

		const organizationName = data.get('organizationName')?.toString();
		const projectName = data.get('projectName')?.toString();

		if (databaseId === '0' || databaseId === undefined) {
			return { error: 'Database is required' };
		}
		if (organizationName === '' || organizationName === undefined) {
			return { error: 'Organization name is required' };
		}
		if (projectName === '' || projectName === undefined) {
			return { error: 'Project name is required' };
		}

		const returned = await client
			.DELETE('/redis/{id}', {
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
