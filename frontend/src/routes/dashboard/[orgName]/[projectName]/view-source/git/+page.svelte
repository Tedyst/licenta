<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;
</script>

<div class="hero bg-base-200">
	<div class="hero-content flex-col">
		<div class="text-center lg:text-left">
			<p class="py-6">
				Here you can view secrets found from this Git repository. Most of these are usually false
				positives, and are harmless.
			</p>
		</div>
		{data.gitRepo?.git_repository}

		{#each data?.commits || [] as commit}
			<div class={'collapse ' + (commit.results.length ? 'bg-error text-white' : 'bg-base-300')}>
				<input type="radio" name="my-accordion-1" checked />
				<div class="collapse-title overflow-hidden">
					<div class="flex justify-between items-center">
						<div class="text-xl font-medium">{commit.author}</div>
						<div class="text-base">{commit.commit_date}</div>
					</div>
					<div class="text-sm">{commit.description}</div>
					<div class="text-xs font-very-small">{commit.commit_hash}</div>
				</div>
				<div class="collapse-content gap-2 flex flex-col text-black">
					{#each commit?.results as result}
						<div class="card bg-slate-200">
							<div class="card-body">
								<div class="card-title">
									Offending Secret: {result.name}
								</div>
								<p class="text-gray-500 break-all"><b>File:</b> {result.filename}</p>
								<p class="text-gray-500 break-all"><b>Line:</b> {result.line}</p>
								<p class="text-gray-500 break-all">
									<b>Line Number:</b>
									{result.line_number}
								</p>
								<p class="text-gray-500 break-all">
									<b>Username:</b>
									{result.username}
								</p>
								<p class="text-gray-500 break-all">
									<b>Password:</b>
									{result.password}
								</p>
							</div>
						</div>
					{/each}
					{#if commit.results.length === 0}
						<div class="card bg-base-300">
							<div class="card-body">
								<div class="card-title">No secrets found in this commit.</div>
							</div>
						</div>
					{/if}
				</div>
			</div>
		{/each}
	</div>
</div>
