<script lang="ts">
	import { onMount } from 'svelte';
	import { unfulfilledOrderSummary, type UnfulfilledOrderSummary } from '$lib/api';
	import Button from '$lib/Button.svelte';
	import { constructTables } from './summary';

	let summary: UnfulfilledOrderSummary | undefined = $state(undefined);
	const refresh = async () => {
		summary = await unfulfilledOrderSummary();
	};
	onMount(refresh);

	const MAX_SAMPLE = 10;

	const summaryTables = $derived.by(() => {
		if (!summary) return [];
		return constructTables(summary);
	});
</script>

<div class="flex flex-col gap-4">
	<div class="flex">
		<Button onClick={refresh}>Refresh</Button>
	</div>
	<div class="flex flex-col gap-1">
		{#if summary}
			<h2 class="text-xl">Unfulfilled Order IDs</h2>
			<div class="flex gap-2">
				{#each summary.order_id_samples as order_id}
					<span>{order_id}</span>
				{/each}
				{#if summary.order_id_samples.length >= MAX_SAMPLE}
					<span>(and more)</span>
				{/if}
			</div>
		{/if}
	</div>
	<div class="flex flex-col gap-1">
		<h2 class="text-xl">Unfulfilled Items</h2>
		{#each summaryTables as tbl}
			<h3 class="text-lg">{tbl.name}</h3>
			<table>
				<thead>
					{#each tbl.columns as colRow}
						<tr>
							<th></th>
							{#each colRow as col}
								<th colspan={col.span}>{col.label}</th>
							{/each}
						</tr>
					{/each}
				</thead>
				<tbody>
					{#each tbl.rows as row}
						<tr>
							<td class="font-bold">{row.label}</td>
							{#each row.data as num}
								<td>
									{#if num}{num}{:else}-{/if}
								</td>
							{/each}
						</tr>
					{/each}
				</tbody>
			</table>
		{/each}
	</div>
</div>

<style lang="postcss">
	table {
		@apply w-fit;
	}
	td,
	th {
		@apply border border-black px-2 text-center;
	}
	tbody tr:nth-child(2n + 1) {
		@apply bg-gray-300;
	}
</style>
