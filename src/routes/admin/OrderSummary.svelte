<script lang="ts">
	import { onMount } from 'svelte'
	import { orderSummary, type OrderSummary } from '$lib/api'
	import Button from '$lib/Button.svelte'
	import { constructTables } from './summary'
	import ErrorBoundary from '$lib/ErrorBoundary.svelte'

	interface Props {
		searchOrder: (phrase: string) => void
	}
	const { searchOrder }: Props = $props()

	let showCollected = $state(false)
	let error: unknown = $state()
	let summary: OrderSummary | undefined = $state(undefined)
	const refresh = async () => {
		try {
			summary = await orderSummary(showCollected)
		} catch (e) {
			error = e
		}
	}
	onMount(refresh)
	$effect(() => {
		const _ = showCollected
		refresh()
	})

	const MAX_SAMPLE = 10

	const search = (label: string) => () => searchOrder(label)
	const summaryTables = $derived.by(() => {
		if (!summary) return []
		return constructTables(summary)
	})
</script>

<div class="flex flex-col gap-4">
	<div class="flex items-center gap-4">
		<Button onClick={refresh}>Refresh</Button>
		<label class="flex gap-2">
			<input type="checkbox" bind:checked={showCollected} />
			Show Collected Orders
		</label>
	</div>
	<ErrorBoundary {error} />
	<div class="flex flex-col gap-1">
		{#if summary}
			<h2 class="text-xl">Unfulfilled Order IDs</h2>
			<div class="flex flex-wrap gap-x-2">
				{#each summary.order_id_samples as order_id}
					<button onclick={search(order_id)} class="text-blue-800 underline">
						{order_id}
					</button>
				{/each}
				{#if summary.order_id_samples.length >= MAX_SAMPLE}
					<span>(and more)</span>
				{/if}
			</div>
		{/if}
	</div>
	<div class="flex flex-col gap-1">
		<h2 class="text-xl">
			{#if !showCollected}Unfulfilled
			{/if} Items
		</h2>
		{#each summaryTables as tbl}
			<h3 class="mt-4 text-lg">{tbl.name}</h3>
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
							{#each row.data as num, i}
								<td onclick={search(row.fullLabels[i])}>
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
