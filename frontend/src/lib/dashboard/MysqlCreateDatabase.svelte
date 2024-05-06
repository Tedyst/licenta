<script lang="ts">
	export let project: components['schemas']['Project'];

	import client from '$lib/client';

	import { toast } from 'svelte-daisy-toast';
	import BaseCreateDatabase from './BaseCreateDatabase.svelte';
	import type { Field } from './BaseCreateDatabase.svelte';
	import type { components } from '$lib/api/v1';
	import { invalidate } from '$app/navigation';

	let createAction = (data: Record<Field, string>) => {
		if (!project) return;

		client
			.POST('/mysql', {
				body: {
					database_name: data.database,
					host: data.hostname,
					password: data.password,
					port: parseInt(data.port),
					project_id: project?.id,
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
				invalidate('app:mysql');
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
