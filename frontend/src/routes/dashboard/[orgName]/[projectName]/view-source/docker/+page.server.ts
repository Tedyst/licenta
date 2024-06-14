import type { PageServerLoad } from './$types';
import { clientFromFetch } from '$lib/client';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ fetch, parent, url }) => {
	const parentData = await parent();

	const dockerId = url.searchParams.get('id') || '0';
	if (!dockerId) {
		redirect(302, '/dashboard/' + parentData.organization?.name + '/' + parentData.project?.name);
	}

	const client = clientFromFetch(fetch);

	const returned = await client
		.GET('/docker/{id}', { params: { path: { id: +dockerId } } })
		.then((res) => {
			if (res.data?.success) {
				return {
					image: res.data.image,
					layers: res.data.layers
				};
			}
			return { error: res.error?.message || 'An error occurred' };
		});

	if (returned?.error) {
		return { error: returned.error };
	}

	return {
		...parentData,
		image: returned.image,
		layers: returned.layers
	};
};
