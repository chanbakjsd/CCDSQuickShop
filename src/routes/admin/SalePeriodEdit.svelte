<script lang="ts">
	import { onMount } from 'svelte'

	import { salePeriods, updateSalePeriod, type SalePeriod } from '$lib/api'
	import { formatDate } from '$lib/util'
	import ErrorBoundary from '$lib/ErrorBoundary.svelte'
	import Options from '$lib/Options.svelte'
	import MerchEdit from './MerchEdit.svelte'
	import OrderSummaryView from './OrderSummaryView.svelte'
	import CouponEdit from './CouponEdit.svelte'
	import Button from '$lib/Button.svelte'

	interface Props {
		searchOrder: (phrase: string) => void
	}
	const { searchOrder }: Props = $props()

	let selected: string | undefined = $state(undefined)
	const options = [{ text: 'Merch' }, { text: 'Coupons' }, { text: 'Order Summary' }]

	let loading = $state(true)
	let viewing: string | undefined = $state(undefined)
	let editing: number | undefined = $state(undefined)
	let selectedStartDate = $state('')
	let periods: SalePeriod[] = $state([])
	onMount(() => {
		salePeriods().then((p) => {
			periods = p
			loading = false
		})
	})

	const goBack = () => {
		viewing = undefined
	}
	const view = (id: string) => () => {
		viewing = id
		editing = undefined
	}
	const edit = (idx: number) => () => {
		viewing = undefined
		editing = idx
		selectedStartDate = formatDate(periods[idx].start_time, 'T')
	}

	let updateError: unknown = $state()
	const updatePeriod = async () => {
		if (editing === undefined) {
			throw new Error('expected editing to be non-undefined')
		}
		const newPeriod = {
			...periods[editing],
			start_time: new Date(selectedStartDate)
		}
		try {
			periods[editing] = await updateSalePeriod(newPeriod)
		} catch (e) {
			updateError = e
		}
	}

	$effect(() => {
		if (!loading && periods[periods.length - 1].id !== '') {
			periods = [
				...periods,
				{
					id: '',
					start_time: new Date(),
					name: '(New Entry)'
				}
			]
		}
	})
</script>

{#if viewing}
	<div class="flex items-center gap-4">
		<Button onClick={goBack} size="md">Back</Button>
		Viewing {periods.find((x) => x.id === viewing)?.name} ({viewing}).
	</div>
	<Options {options} bind:value={selected} />
	{#if selected === 'Merch'}
		<MerchEdit salePeriod={viewing} />
	{:else if selected === 'Coupons'}
		<CouponEdit salePeriod={viewing} />
	{:else if selected === 'Order Summary'}
		<OrderSummaryView salePeriod={viewing} {searchOrder} />
	{/if}
{:else}
	<table class="w-fit border border-black text-center">
		<thead>
			<tr>
				<th>#</th>
				<th>Name</th>
				<th>Start Time</th>
				<th></th>
			</tr>
		</thead>
		<tbody>
			{#each periods as period, i}
				<tr class="odd:bg-gray-200">
					<td>{period.id}</td>
					<td>{period.name}</td>
					<td>{formatDate(period.start_time)}</td>
					<td class="flex gap-1">
						{#if period.id}<Button onClick={view(period.id)} size="md">View</Button>{/if}
						<Button onClick={edit(i)} size="md">Edit</Button>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>

	{#if editing !== undefined}
		<div class="grid grid-cols-[auto,1fr] gap-x-2 gap-y-1">
			<span>ID</span>
			<input disabled value={periods[editing].id} placeholder="Sale Period ID (auto-generated)" />
			<span>Name</span>
			<input bind:value={periods[editing].name} placeholder="Name of the Sale Period" />
			<span>Start Time</span>
			<input type="datetime-local" bind:value={selectedStartDate} />
		</div>
		<div class="flex">
			<Button onClick={updatePeriod}>Update Period</Button>
			<ErrorBoundary error={updateError} />
		</div>
	{/if}
{/if}

<style lang="postcss">
	th,
	td {
		@apply px-4 py-1;
	}
	input {
		@apply max-w-64 border border-black px-1;
	}
</style>
