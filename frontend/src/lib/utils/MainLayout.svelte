<script lang="ts">
	import { user, organizations } from '$lib/stores';
	import { theme } from '../../lib/stores';
	import GravatarImage from '$lib/utils/GravatarImage.svelte';

	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import client, { updateCurrentUser, updateOrganizations } from '../../lib/client';

	import { pa } from '@accuser/svelte-plausible-analytics';
	import ListOrganizationsAndProjects from '$lib/utils/ListOrganizationsAndProjects.svelte';
	const { addEvent } = pa;

	function toggleTheme() {
		addEvent('theme-toggle');
		theme.set($theme === 'dark' ? 'light' : 'dark');
	}

	let serverError = '';
	onMount(async () => {
		if ($user) {
			return;
		}

		let results = await Promise.all([updateCurrentUser(), updateOrganizations()]);
		if (results[0]) {
			serverError = results[0];
			return;
		}
		if (results[1]) {
			serverError = results[1];
			return;
		}
	});
</script>

<div class="drawer md:drawer-open">
	<input id="my-drawer" type="checkbox" class="drawer-toggle" />
	<div class="drawer-content flex flex-col">
		<div class="w-full navbar bg-base-300">
			<div class="flex-none md:hidden">
				<label for="my-drawer" aria-label="open sidebar" class="btn btn-square btn-ghost">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						class="inline-block w-6 h-6 stroke-current"
						><path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M4 6h16M4 12h16M4 18h16"
						/></svg
					>
				</label>
			</div>
			<div class="flex-none block grow">
				<ul class="menu menu-horizontal hidden md:inline-block">
					<li><a href="/organizations">Organizations</a></li>
				</ul>
				<ul class="menu menu-horizontal hidden md:inline-block">
					<li><a href="/projects">Projects</a></li>
				</ul>
				{#if $user?.admin}
					<ul class="menu menu-horizontal hidden md:inline-block">
						<li><a href="/admin">Admin</a></li>
					</ul>
				{/if}
			</div>

			<label class="swap swap-rotate mr-5">
				<input
					id="toggle-theme"
					type="checkbox"
					on:click={toggleTheme}
					checked={$theme !== 'dark'}
					aria-label="Toggle Theme"
				/>

				<script>
					(function () {
						let localTheme = localStorage.getItem('theme');
						if (localTheme) {
							if (typeof document === 'undefined') return;
							document.getElementById('toggle-theme').checked = JSON.parse(localTheme) !== 'dark';
						}
					})();
				</script>

				<svg
					class="swap-on fill-current w-8 h-8"
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 24 24"
					><path
						d="M5.64,17l-.71.71a1,1,0,0,0,0,1.41,1,1,0,0,0,1.41,0l.71-.71A1,1,0,0,0,5.64,17ZM5,12a1,1,0,0,0-1-1H3a1,1,0,0,0,0,2H4A1,1,0,0,0,5,12Zm7-7a1,1,0,0,0,1-1V3a1,1,0,0,0-2,0V4A1,1,0,0,0,12,5ZM5.64,7.05a1,1,0,0,0,.7.29,1,1,0,0,0,.71-.29,1,1,0,0,0,0-1.41l-.71-.71A1,1,0,0,0,4.93,6.34Zm12,.29a1,1,0,0,0,.7-.29l.71-.71a1,1,0,1,0-1.41-1.41L17,5.64a1,1,0,0,0,0,1.41A1,1,0,0,0,17.66,7.34ZM21,11H20a1,1,0,0,0,0,2h1a1,1,0,0,0,0-2Zm-9,8a1,1,0,0,0-1,1v1a1,1,0,0,0,2,0V20A1,1,0,0,0,12,19ZM18.36,17A1,1,0,0,0,17,18.36l.71.71a1,1,0,0,0,1.41,0,1,1,0,0,0,0-1.41ZM12,6.5A5.5,5.5,0,1,0,17.5,12,5.51,5.51,0,0,0,12,6.5Zm0,9A3.5,3.5,0,1,1,15.5,12,3.5,3.5,0,0,1,12,15.5Z"
					/></svg
				>
				<svg
					class="swap-off fill-current w-8 h-8"
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 24 24"
					><path
						d="M21.64,13a1,1,0,0,0-1.05-.14,8.05,8.05,0,0,1-3.37.73A8.15,8.15,0,0,1,9.08,5.49a8.59,8.59,0,0,1,.25-2A1,1,0,0,0,8,2.36,10.14,10.14,0,1,0,22,14.05,1,1,0,0,0,21.64,13Zm-9.5,6.69A8.14,8.14,0,0,1,7.08,5.22v.27A10.15,10.15,0,0,0,17.22,15.63a9.79,9.79,0,0,0,2.1-.22A8.11,8.11,0,0,1,12.14,19.73Z"
					/></svg
				>
			</label>
			<div class="dropdown dropdown-hover dropdown-end">
				<div class="avatar">
					<div tabindex="0" role="button" class="w-10 rounded-full">
						<GravatarImage email={$user?.email} />
					</div>
				</div>
				<ul
					tabindex="-1"
					class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-32"
				>
					<li><a href="/account">Settings</a></li>
					<li class="divider m-0 flex-nowrap shrink-0 opacity-100 bg-inherit" />
					<li><a href="/logout">Logout</a></li>
				</ul>
			</div>
		</div>
		<div class="m-4">
			{#if serverError !== ''}
				<div role="alert" class="alert alert-error mb-4">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="stroke-current shrink-0 h-6 w-6"
						fill="none"
						viewBox="0 0 24 24"
						><path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
						/></svg
					>
					<span>Error! Cannot access server: {serverError}</span>
				</div>
			{/if}
			<slot />
		</div>
	</div>
	<div class="drawer-side">
		<label for="my-drawer" aria-label="close sidebar" class="drawer-overlay" />
		<ul class="menu p-4 w-52 min-h-full bg-base-300">
			<li class="mb-4 font-semibold text-xl">
				<a href="/dashboard">Licenta</a>
			</li>

			<li class="mb-4 md:hidden"><a href="/dashboard">Organizations</a></li>
			<li class="mb-4 md:hidden"><a href="/dashboard">Projects</a></li>
			{#if $user?.admin}
				<li class="mb-4 md:hidden"><a href="/dashboard">Admin</a></li>
			{/if}

			<li class="divider md:hidden m-0 flex-nowrap shrink-0 opacity-100 bg-inherit" />

			<ListOrganizationsAndProjects />
		</ul>
	</div>
</div>
