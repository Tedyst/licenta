<script lang="ts">
	import { Md5 } from 'ts-md5';

	export let email: string | undefined;
	export let size = 100;

	let hash = '';
	let element: HTMLImageElement;

	$: hash = Md5.hashStr(email?.toLowerCase()?.trim() || '');
	$: if (size < 1) size = 1;
	$: if (size > 2048) size = 2048;
	$: if (element) element.src = `https://www.gravatar.com/avatar/${hash}?s=${size}&d=mp`;
</script>

{#if !email}
	<div class="skeleton w-16 h-16 rounded-full shrink-0" />
{:else}
	<img id="gravatar" alt="gravatar" bind:this={element} />
{/if}
