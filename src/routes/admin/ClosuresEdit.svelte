<script lang="ts">
	import { onMount } from 'svelte';
	import { listStoreClosures, updateStoreClosure, type StoreClosure } from '$lib/api';
	import { formatDate } from '$lib/util';
	import Button from '$lib/Button.svelte';

	let closures: StoreClosure[];
	let loading = true;
	onMount(() => {
		listStoreClosures().then((x) => {
			closures = x;
			loading = false;
		});
	});

	$: {
		if (!loading && (closures.length === 0 || closures[closures.length - 1].id !== '')) {
			closures.push({
				id: '',
				start_time: new Date(),
				end_time: new Date(9999, 11, 31),
				message: '(New Entry)',
				show_order_check: true
			});
		}
	}

	let selected = -1;
	let selectedStartDate = '';
	let selectedEndDate = '';
	const select = (i: number) => () => {
		selected = i;
		selectedStartDate = formatDate(closures[selected].start_time, 'T');
		selectedEndDate = formatDate(closures[selected].end_time, 'T');
	};

	const updateClosure = async () => {
		const newClosure = {
			...closures[selected],
			start_time: new Date(selectedStartDate),
			end_time: new Date(selectedEndDate)
		};
		closures[selected] = await updateStoreClosure(newClosure);
	};
</script>

<table class="w-fit border border-black text-center">
	<thead>
		<tr>
			<th>#</th>
			<th>Start</th>
			<th>End</th>
			<th>Message</th>
			<th>Show Order?</th>
		</tr>
	</thead>
	<tbody>
		{#each closures as closure, i}
			<tr class="odd:bg-gray-200" on:click={select(i)}>
				<td>{closure.id}</td>
				<td>{formatDate(closure.start_time)}</td>
				<td>{formatDate(closure.end_time)}</td>
				<td>{closure.message}</td>
				<td>{closure.show_order_check ? 'Yes' : 'NO'}</td>
			</tr>
		{/each}
	</tbody>
</table>

{#if selected !== -1}
	<div class="grid grid-cols-[auto,1fr] gap-x-2 gap-y-1">
		<span>ID</span>
		<input disabled value={closures[selected].id} placeholder="Closure ID (auto-generated)" />
		<span>Start</span>
		<input type="datetime-local" bind:value={selectedStartDate} />
		<span>End</span>
		<input type="datetime-local" bind:value={selectedEndDate} />
		<span>Message</span>
		<input bind:value={closures[selected].message} />
		<span>Show Order</span>
		<input type="checkbox" bind:checked={closures[selected].show_order_check} />
		<div class="flex"><Button onClick={updateClosure}>Update Closure</Button></div>
	</div>
{/if}

<style lang="postcss">
	th,
	td {
		@apply px-4 py-1;
	}
	input {
		@apply max-w-64 border border-black px-1;
	}
	input[type='checkbox'] {
		@apply justify-self-start;
		max-width: unset;
	}
</style>
