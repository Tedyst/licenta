import createClient from 'openapi-fetch';
import type { LayoutServerLoad } from './$types';
import type { components, paths } from '$lib/api/v1';
import { goto } from '$app/navigation';
export const prerender = false;
export const ssr = true;
import { env } from '$env/dynamic/public';

export const load: LayoutServerLoad = async ({ fetch, depends }) => {
	const client = createClient<paths>({
		baseUrl: env.PUBLIC_BACKEND_URL + '/api/v1',
		fetch: fetch,
		credentials: 'include'
	});

	depends('app:userinfo', 'app:organizationinfo');

	const userInfo = client
		.GET('/users/me')
		.then((res) => {
			if (res.data?.success) {
				return {
					user: res.data.user
				};
			} else if (res.data?.success === false) {
				goto('/login');
				return {
					user: null,
					error: 'Not logged in'
				};
			}
			return {
				user: null,
				error: res.error?.message || 'Internal server error'
			};
		})
		.catch((err) => {
			return {
				user: null,
				error: err.message
			};
		});
	const organizationInfo = client
		.GET('/organizations')
		.then((res) => {
			if (res.data?.success) {
				return {
					organizations: res.data.organizations
				};
			}
			return {
				organizations: [],
				error: res.error?.message || 'Internal server error'
			};
		})
		.catch((err) => {
			return {
				organizations: [],
				error: err.message
			};
		});

	const promises = await Promise.all([userInfo, organizationInfo]);

	return {
		user: promises[0].user as components['schemas']['User'],
		organizations: promises[1].organizations as components['schemas']['Organization'][],
		error: promises[0].error || promises[1].error
	};
};
