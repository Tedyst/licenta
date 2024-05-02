<script lang="ts">
	import client from '$lib/client';
	import { currentProject, currentPostgresDatabases } from '$lib/stores';

	import { toast } from 'svelte-daisy-toast';
	import BaseCreateDatabase from './BaseCreateDatabase.svelte';
	import type { Field } from './BaseCreateDatabase.svelte';

	let createAction = (data: Record<Field, string>) => {
		if (!$currentProject) return;

		client
			.POST('/postgres', {
				body: {
					database_name: data.database,
					host: data.hostname,
					password: data.password,
					port: parseInt(data.port),
					project_id: $currentProject?.id,
					username: data.username
				}
			})
			.then((response) => {
				if (response.error) {
					toast({
						closable: true,
						duration: 5000,
						message: response.error.message,
						type: 'error'
					});
					return;
				}
				$currentPostgresDatabases = [
					...$currentPostgresDatabases,
					response.data?.postgres_database
				];
				toast({
					closable: true,
					duration: 5000,
					message: 'Database created successfully',
					type: 'success'
				});
			});
	};
</script>

<BaseCreateDatabase {createAction} />
