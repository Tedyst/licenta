import createClient from 'openapi-fetch';
import type { LayoutServerLoad } from './$types';
import type { paths } from '$lib/api/v1';

export const load: LayoutServerLoad = async ({ params, parent, fetch, url, depends }) => {
	const parentData = await parent();

	const currentProject =
		parentData.organization?.projects.filter((v) => v.name == params.projectName).at(0) || null;

	const client = createClient<paths>({
		baseUrl: url.origin + '/api/v1',
		fetch: fetch,
		credentials: 'include'
	});

	if (!currentProject) {
		return {
			...parentData
		};
	}
	const mysqlDatabases = client.GET('/mysql', {
		params: { query: { project: currentProject.id } }
	});
	const postgresDatabases = client.GET('/postgres', {
		params: { query: { project: currentProject.id } }
	});

	depends('app:mysql', 'app:postgres');

	const promises = await Promise.all([mysqlDatabases, postgresDatabases]);

	return {
		...parentData,
		project: currentProject,
		mysqlDatabases: promises[0].data?.mysql_databases,
		postgresDatabases: promises[1].data?.postgres_databases
	};
};
