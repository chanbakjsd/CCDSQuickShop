<script lang="ts">
	import type { Order } from '$lib/api';
	import Button from '$lib/Button.svelte';
	import IconChevronDown from '$lib/IconChevronDown.svelte';
	import IconChevronUp from '$lib/IconChevronUp.svelte';
	import Invoice from '$lib/Invoice.svelte';

	export let order: Order;
	export let expanded = false;
	export let userFacing = false;
	export let collect: (() => any) | null = null;
	export let cancel: (() => any) | null = null;

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

	$: orderStatus =
		order.collectionTime !== null
			? ('COLLECTED' as const)
			: order.cancelled
				? ('CANCELLED' as const)
				: order.paymentTime === null
					? ('PAYMENT PENDING' as const)
					: ('UNCOLLECTED' as const);

	const pillColors = {
		COLLECTED: 'bg-gray-400',
		CANCELLED: 'bg-red-400',
		'PAYMENT PENDING': 'bg-red-400',
		UNCOLLECTED: null
	} as const;
	const messages = {
		COLLECTED: {
			text: 'Your merch has been collected. Enjoy your merch!',
			css: 'border-gray-400 bg-gray-100'
		},
		CANCELLED: {
			text: 'Your order has been cancelled. This is likely due to non-payment.\nPlease contact SCDS Club if this is a mistake.',
			css: 'border-red-400 bg-red-100'
		},
		'PAYMENT PENDING': {
			text: 'We are waiting for confirmation from our payment provider.\nTry refreshing after a few moments if you have already paid.',
			css: 'border-yellow-400 bg-yellow-100'
		},
		UNCOLLECTED: {
			text: 'Your order has been confirmed!\nWe will send details on merch collection to your NTU email once they are ready.',
			css: 'border-green-400 bg-green-100'
		}
	} as const;

	$: bgColor = orderStatus === 'UNCOLLECTED' ? 'bg-white' : 'bg-gray-200';
	$: pillColor = pillColors[orderStatus];
	$: message = messages[orderStatus];
</script>

<div class={`w-full shadow-md lg:w-1/2 ${bgColor}`}>
	<button
		class="flex w-full items-center justify-between p-4 text-xl font-bold"
		on:click={toggleExpanded}
	>
		<div class="flex items-center gap-2">
			<span>Order {order.id}</span>
			{#if pillColor && !userFacing}<span class={`pill ${pillColor}`}>{orderStatus}</span>{/if}
		</div>
		{#if expanded}<IconChevronUp />{:else}<IconChevronDown />{/if}
	</button>
	<div class="overflow-hidden bg-white transition-all" style:max-height={maxHeight}>
		<hr />
		<div class="flex flex-col gap-4 p-4" bind:this={content}>
			{#if userFacing}
				<div
					class={`w-full whitespace-pre rounded-lg border px-4 py-2 text-justify ${message.css}`}
				>
					{message.text}
				</div>
			{/if}
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
					<Button onClick={cancel}>Cancel</Button>
				{/if}
				{#if order.collectionTime === null && collect}
					<Button onClick={collect}>Mark as Collected</Button>
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
