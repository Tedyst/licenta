import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { clientFromFetch } from '$lib/client';

export const load: PageServerLoad = async ({ parent, url, depends }) => {
	const parentData = await parent();
	const client = clientFromFetch(fetch, url.origin);

	const scanId = url.searchParams.get('id') || '0';

	if (scanId === '0') {
		redirect(302, '/dashboard/' + parentData.organization?.name + '/' + parentData.project?.name);
	}

	const currentScan = await client
		.GET('/scan/{id}', {
			params: { path: { id: +scanId } }
		})
		.then((res) => {
			if (res.data?.success) {
				return res.data;
			}
			return null;
		});

	depends('app:currentScan');

	if (!currentScan) {
		redirect(302, '/dashboard/' + parentData.organization?.name + '/' + parentData.project?.name);
	}

	return {
		...parentData,
		scan: {
			scan: currentScan.scan,
			results: currentScan.results,
			bruteforceResults: currentScan.bruteforce_results
		}
	};
};
