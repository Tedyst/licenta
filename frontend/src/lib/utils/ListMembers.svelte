<script lang="ts">
	import type { components } from '$lib/api/v1';
	import GravatarImage from './GravatarImage.svelte';

	export let organization: components['schemas']['Organization'];
</script>

<div class="flex flex-col flex-1">
	<div class="overflow-x-auto">
		<table class="table">
			<thead>
				<tr>
					<th>Name</th>
					<th>Email</th>
					<th>Role</th>
				</tr>
			</thead>
			<tbody>
				{#each organization.members as member}
					<tr>
						<td>
							<div class="flex items-center gap-3">
								<div class="avatar">
									<div class="mask mask-squircle w-12 h-12">
										<GravatarImage email={member.email} size={56} />
									</div>
								</div>
								<div>
									<div>
										{member.username}
									</div>
								</div>
							</div>
						</td>
						<td>{member.email}</td>
						<td>{member.role}</td>
						<td>
							<a
								href="/dashboard/{organization.name}/edit-user?userId={member.id}"
								class="btn btn-secondary"
							>
								Edit Role
							</a>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	<div class="flex justify-center">
		<a href="/dashboard/{organization.name}/add-user" class="btn btn-primary mt-4"
			>Add New User to Organization</a
		>
	</div>
</div>
