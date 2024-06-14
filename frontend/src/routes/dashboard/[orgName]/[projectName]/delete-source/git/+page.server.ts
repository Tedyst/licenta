import type { Actions, PageServerLoad } from './$types';
import { clientFromFetch } from '$lib/client';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ parent, url }) => {
	const parentData = await parent();

	const databaseId = url.searchParams.get('id') || '0';

	if (databaseId === '0') {
		redirect(302, '/dashboard/' + parentData.organization?.name + '/' + parentData.project?.name);
	}

	const currentSource =
		parentData.gitRepositories?.filter((v) => v.id === +databaseId).at(0) || null;

	if (!currentSource) {
		redirect(302, '/dashboard/' + parentData.organization?.name + '/' + parentData.project?.name);
	}

	return {
		...parentData,
		currentSource: currentSource
	};
};

export const actions = {
	default: async ({ request, fetch }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch);

		const sourceId = data.get('databaseId')?.toString();

		const organizationName = data.get('organizationName')?.toString();
		const projectName = data.get('projectName')?.toString();

		if (sourceId === '0' || sourceId === undefined) {
			return { error: 'Database is required' };
		}
		if (organizationName === '' || organizationName === undefined) {
			return { error: 'Organization name is required' };
		}
		if (projectName === '' || projectName === undefined) {
			return { error: 'Project name is required' };
		}

		const returned = await client
			.DELETE('/git/{id}', {
				params: { path: { id: +sourceId } }
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
