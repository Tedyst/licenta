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
	const redisDatabases = client.GET('/redis', {
		params: { query: { project: currentProject.id } }
	});
	const mongoDatabases = client.GET('/mongo', {
		params: { query: { project: currentProject.id } }
	});
	const gitRepositories = client.GET('/git', {
		params: { query: { project: currentProject.id } }
	});
	const dockerImages = client.GET('/docker', {
		params: { query: { project: currentProject.id } }
	});
	const scanGroups = client.GET('/scan-groups', {
		params: { query: { project: currentProject.id } }
	});

	depends('app:mysql', 'app:postgres', 'app:git', 'app:docker', 'app:scan-groups');

	const promises = await Promise.all([
		mysqlDatabases,
		postgresDatabases,
		gitRepositories,
		dockerImages,
		scanGroups,
		mongoDatabases,
		redisDatabases
	]);

	return {
		...parentData,
		project: currentProject,
		mysqlDatabases: promises[0].data?.mysql_databases,
		postgresDatabases: promises[1].data?.postgres_databases,
		gitRepositories: promises[2].data?.git_repositories,
		dockerImages: promises[3].data?.images,
		scanGroups: promises[4].data?.scan_groups,
		redisDatabases: promises[6].data?.redis_databases,
		mongoDatabases: promises[5].data?.mongo_databases
	};
};
