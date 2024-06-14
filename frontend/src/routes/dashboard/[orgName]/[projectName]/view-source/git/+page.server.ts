import type { PageServerLoad } from './$types';
import { clientFromFetch } from '$lib/client';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ fetch, parent, url }) => {
	const parentData = await parent();

	const gitId = url.searchParams.get('id') || '0';
	if (!gitId) {
		redirect(302, '/dashboard/' + parentData.organization?.name + '/' + parentData.project?.name);
	}

	const client = clientFromFetch(fetch);

	const returned = await client
		.GET('/git/{id}', { params: { path: { id: +gitId } } })
		.then((res) => {
			if (res.data?.success) {
				return {
					gitRepo: res.data.git,
					commits: res.data.commits
				};
			}
			return { error: res.error?.message || 'An error occurred' };
		});

	if (returned?.error) {
		return { error: returned.error };
	}

	return {
		...parentData,
		gitRepo: returned.gitRepo,
		commits: returned.commits
	};
};
