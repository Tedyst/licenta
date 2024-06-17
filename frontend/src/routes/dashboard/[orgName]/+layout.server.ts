import { clientFromFetch } from '$lib/client';
import type { LayoutServerLoad } from './$types';

export const prerender = false;

export const load: LayoutServerLoad = async ({ params, parent, fetch }) => {
	const parentData = await parent();

	const currentOrganization =
		parentData.organizations?.filter((v) => v.name == params.orgName).at(0) || null;

	const client = clientFromFetch(fetch);

	const workers = await client.GET('/worker', {
		params: { query: { organization: currentOrganization?.id || 0 } }
	});

	console.log(workers);

	return {
		...parentData,
		organization: currentOrganization,
		workers: workers.data?.workers || []
	};
};
