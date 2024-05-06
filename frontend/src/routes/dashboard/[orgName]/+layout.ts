import type { LayoutLoad } from './$types';

export const prerender = false;

export const load: LayoutLoad = async ({ params, parent }) => {
	const parentData = await parent();
	return {
		...parentData,
		organization: parentData.organizations?.filter((v) => v.name == params.orgName).at(0) || null
	};
};
