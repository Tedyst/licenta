<script lang="ts">
	export let project: components['schemas']['Project'] | null;
	import client from '$lib/client';

	import { toast } from 'svelte-daisy-toast';
	import BaseCreateDatabase from './BaseCreateDatabase.svelte';
	import type { Field } from './BaseCreateDatabase.svelte';
	import { invalidate } from '$app/navigation';
	import type { components } from '$lib/api/v1';

	let createAction = (data: Record<Field, string>) => {
		if (!project) return;

		client
			.POST('/postgres', {
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
				invalidate('app:postgres');
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
