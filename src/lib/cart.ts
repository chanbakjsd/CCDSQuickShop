import { z } from 'zod';

export const formatPrice = (price: number) => price.toFixed(2)

export type CartItem = {
	id: string
	name: string
	variant: {
		type: string
		option: string
	}[]
	imageURL: string
	amount: number
	unitPrice: number // The unit price of the item in cents.
}

export const Requirement = z.object({
	type: z.literal("purchase_count"),
	amount: z.number(),
})

// Discount is used for previewing discounts in the cart.
export const Discount = z.object({
	type: z.literal("percentage"),
	amount: z.number(), // 40% discount (i.e. pay 60%) is represented as 40.
})

export const Coupon = z.object({
	requirements: Requirement.array(),
	couponCode: z.string(),
	discount: Discount,
});

export type Requirement = z.infer<typeof Requirement>
export type Discount = z.infer<typeof Discount>
export type Coupon = z.infer<typeof Coupon>

export const calculateCartTotal = (cart: CartItem[]) =>
	cart.reduce<number>((total, item) => total + item.unitPrice * item.amount, 0);

export const applyCoupon = (cart: CartItem[], coupon: Coupon) => {
	const cartTotal = calculateCartTotal(cart);
	switch (coupon.discount.type) {
		case 'percentage':
			return Math.ceil((cartTotal * (100 - coupon.discount.amount)) / 100);
	}
};

export const checkRequirement = (cart: CartItem[], requirement: Requirement) => {
	switch (requirement.type) {
		case 'purchase_count':
			return cart.reduce<number>((total, item) => total + item.amount, 0) >= requirement.amount;
	}
};
