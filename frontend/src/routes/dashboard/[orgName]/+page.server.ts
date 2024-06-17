import { clientFromFetch } from '$lib/client';
import type { Actions } from './$types';

export const actions = {
	create_worker: async ({ request, fetch }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch);

		const workerName = data.get('workerName')?.toString();
		const organizationId = data.get('organizationId')?.toString();

		if (organizationId === '0' || organizationId === undefined) {
			return { error: 'Organization name is required' };
		}
		if (workerName === '0' || workerName === undefined) {
			return { error: 'Worker name is required' };
		}

		const returned = await client
			.POST('/worker', {
				body: { name: workerName, organization: +organizationId }
			})
			.then((res) => {
				if (res.response.ok) {
					return {};
				}
				return { error: res.error?.message || 'An error occurred' };
			});

		if (returned.error) {
			return { error: returned.error };
		}
	},
	delete_worker: async ({ request, fetch }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch);

		const workerId = data.get('workerId')?.toString();

		if (workerId === '0' || workerId === undefined) {
			return { error: 'Worker ID is required' };
		}

		const returned = await client
			.DELETE('/worker/{id}', {
				params: { path: { id: +workerId } }
			})
			.then((res) => {
				if (res.response.ok) {
					return {};
				}
				return { error: res.error?.message || 'An error occurred' };
			});

		if (returned.error) {
			return { error: returned.error };
		}
	}
} satisfies Actions;
