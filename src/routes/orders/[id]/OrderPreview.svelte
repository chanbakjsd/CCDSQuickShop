<script lang="ts">
	import type { Order } from '$lib/api';
	import Button from '$lib/Button.svelte';
	import IconChevronDown from '$lib/IconChevronDown.svelte';
	import IconChevronUp from '$lib/IconChevronUp.svelte';
	import Invoice from '$lib/Invoice.svelte';

	export let order: Order;
	export let expanded = false;
	export let collect: (() => void) | null = null;
	export let cancel: (() => void) | null = null;

	let content: HTMLDivElement;
	const toggleExpanded = () => {
		expanded = !expanded;
	};

	$: maxHeight = expanded && content ? `${content.clientHeight}px` : '0px';
	const padZero = (s: number) => '0'.repeat(2 - (s + '').length) + s;
	const formatDate = (date: Date | null) =>
		date
			? `${date.getFullYear()}-${padZero(date.getMonth() + 1)}-${padZero(date.getDate())} ${padZero(date.getHours())}:${padZero(date.getMinutes())}:${padZero(date.getSeconds())}`
			: 'N/A';

	$: bgColor =
		order.collectionTime !== null || order.cancelled || order.paymentTime === null
			? 'bg-gray-200'
			: 'bg-white';
</script>

<div class={`w-full shadow-md lg:w-1/2 ${bgColor}`}>
	<button
		class="flex w-full items-center justify-between p-4 text-xl font-bold"
		on:click={toggleExpanded}
	>
		<div class="flex items-center gap-2">
			<span>Order {order.id}</span>
			{#if order.collectionTime !== null}
				<span class="pill bg-gray-400">COLLECTED</span>
			{:else if order.cancelled}
				<span class="pill bg-red-400">CANCELLED</span>
			{:else if order.paymentTime === null}
				<span class="pill bg-red-400">PAYMENT PENDING</span>
			{/if}
		</div>
		{#if expanded}
			<IconChevronUp />
		{:else}
			<IconChevronDown />
		{/if}
	</button>
	<div class="overflow-hidden bg-white transition-all" style:max-height={maxHeight}>
		<hr />
		<div class="flex flex-col gap-4 p-4" bind:this={content}>
			<div class="grid grid-cols-2 gap-2 xl:grid-cols-3">
				<div class="text-center">
					<div class="font-bold">Name</div>
					<div>{order.name}</div>
				</div>
				<div class="text-center">
					<div class="font-bold">Matric Number</div>
					<div>{order.matricNumber}</div>
				</div>
				<div class="text-center">
					<div class="font-bold">Payment Ref</div>
					<div>{order.paymentRef}</div>
				</div>
				<div class="text-center">
					<div class="font-bold">Payment Time</div>
					<div>{formatDate(order.paymentTime)}</div>
				</div>
				<div class="text-center">
					<div class="font-bold">Collection Time</div>
					<div>{formatDate(order.collectionTime)}</div>
				</div>
				<div class="text-center">
					<div class="font-bold">Email</div>
					<div>{order.email}</div>
				</div>
			</div>
			<hr />
			<Invoice items={order.items} coupon={order.coupon} />
			<div class="flex w-full justify-end gap-2">
				{#if order.paymentTime === null && !order.cancelled && cancel}
					<Button on:click={cancel}>Cancel</Button>
				{/if}
				{#if order.collectionTime === null && collect}
					<Button on:click={collect}>Mark as Collected</Button>
				{/if}
			</div>
		</div>
	</div>
</div>

<style lang="postcss">
	.pill {
		@apply rounded-full px-2 py-1;
	}
</style>
