import { redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { clientFromFetch } from '$lib/client';

export const load: PageServerLoad = async ({ parent, url }) => {
	const parentData = await parent();

	const sourceId = url.searchParams.get('id') || '0';

	if (sourceId === '0') {
		redirect(302, '/dashboard/' + parentData.organization?.name + '/' + parentData.project?.name);
	}

	const currentSource = parentData.dockerImages?.filter((v) => v.id === +sourceId).at(0) || null;

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

		const sourceId = data.get('sourceId')?.toString();
		const image = data.get('image')?.toString();
		const username = data.get('username')?.toString();
		const password = data.get('password')?.toString();

		const organizationName = data.get('organizationName')?.toString();
		const projectName = data.get('projectName')?.toString();

		if (sourceId === '0' || sourceId === undefined) {
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
			.PATCH('/docker/{id}', {
				body: {
					docker_image: image,
					password: password,
					username: username
				},
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
