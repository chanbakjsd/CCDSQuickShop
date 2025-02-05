<script lang="ts">
	import { onMount } from 'svelte';
	import { StoreClosureError } from '$lib/api';
	import Button from '$lib/Button.svelte';

	const timePart = (num: number, unit: string) => {
		num = Math.floor(num) % 60;
		if (num === 1) return `${num} ${unit}`;
		return `${num} ${unit}s`;
	};

	let { data }: { data: StoreClosureError } = $props();
	let time = $state('');
	let interval: number;
	let seconds_left: number | null = $state(data.end_time);
	onMount(() => {
		if (!data.end_time) return;
		interval = setInterval(() => {
			if (!seconds_left) return;
			seconds_left -= 1;
		}, 1000);
	});

	$effect(() => {
		if (seconds_left === null) return;
		switch (true) {
			case seconds_left > 60 * 60:
				time = `${timePart(seconds_left / 60 / 60, 'hour')} and ${timePart(seconds_left / 60, 'minute')} until store opening!`;
				break;
			case seconds_left > 60:
				time = `${timePart(seconds_left / 60, 'minute')} and ${timePart(seconds_left, 'second')} until store opening!`;
				break;
			case seconds_left > 0:
				time = `${timePart(seconds_left, 'second')} until store opening!`;
				break;
			default:
				time = 'Loading...';
				window.location.reload();
				clearInterval(interval);
		}
	});

	let orderID = $state('');
	let orderIDValid = $derived(orderID.match(/^[A-Z]{2}\d{4}$/));
	const visitOrder = () => {
		if (!orderIDValid) return;
		window.location.assign(`/orders/${orderID}`);
	};
</script>

<svelte:head>
	<title>Waiting Room - SCDS Merch Store</title>
</svelte:head>

<div class="flex h-screen w-screen flex-col items-center justify-center gap-2">
	<img
		src="https://ntuscds.com/scse-logo/scds-logo.png"
		class="mb-2 size-32"
		alt="Logo for SCDS Club"
	/>
	<p class="text-xl">{data.user_message}</p>
	{#if time}
		<p class="text-sm">{time}</p>
	{/if}
	{#if data.allow_order}
		<hr class="mb-4 mt-2 w-48 border-brand/50" />
		<p class="text-center text-lg">Have an order ID?</p>
		<div class="flex items-center gap-2">
			<form onsubmit={visitOrder}>
				<input
					placeholder="AB1234"
					class="w-16 border-b border-brand text-center"
					bind:value={orderID}
					maxlength="6"
				/>
			</form>
			<Button size="md" onClick={visitOrder} disabled={!orderIDValid}>Go</Button>
		</div>
	{/if}
</div>
