<script lang="ts">
	import type { PageData } from './$types';
	import ShieldKey from 'svelte-material-icons/ShieldKey.svelte';

	export let data: PageData;
</script>

<h1 class="text-2xl font-bold">Login using a Passkey</h1>
<div class="label flex flex-col justify-center items-center">
	<div class="w-max h-max relative">
		<div class="loading loading-ring w-32 opacity-0" />
		<div class="absolute top-[50%] left-[50%] -translate-x-1/2 -translate-y-1/2">
			<ShieldKey class="text-5xl" />
		</div>
	</div>
</div>

<h1 class="text-sm font-bold">Logging in failed. Please try again.</h1>

<div class="label mt-8 mb-0 pb-0">
	<a href={'/login/webauthn?username=' + data.username} class="label-text-alt link link-hover">
		Click here to go back to try again
	</a>
</div>
<div class="label mt-0 pt-0">
	<a href="/login" class="label-text-alt link link-hover">
		Click here to go back to the login page
	</a>
</div>
