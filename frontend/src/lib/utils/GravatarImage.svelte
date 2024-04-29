<script lang="ts">
	import { Md5 } from 'ts-md5';

	export let email: string | undefined;
	export let size = 100;

	let url: string;

	$: if (email) {
		const hash = Md5.hashStr(email.toLowerCase().trim());
		url = `https://www.gravatar.com/avatar/${hash}?s=${size}&d=mp`;
	}
</script>

{#if !email}
	<div class="skeleton w-16 h-16 rounded-full shrink-0" />
{:else}
	<img id="gravatar" alt="gravatar" src={url} />
{/if}
