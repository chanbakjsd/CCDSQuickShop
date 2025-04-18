import { z } from 'zod';

export const EMAIL_SUFFIX = "@e.ntu.edu.sg"

export const formatPrice = (price: number) => price.toFixed(2)

export const CartItem = z.object({
	id: z.string(),
	name: z.string(),
	variant: z.object({
		type: z.string(),
		option: z.string(),
	}).array(),
	imageURL: z.string(),
	amount: z.number(),
	unitPrice: z.number() // The unit price of the item in cents.
})

export type OrderItem = z.infer<typeof OrderItem>
export const OrderItem = CartItem.omit({
	variant: true
}).extend({
	variant: z.string()
})

export type Item = CartItem | OrderItem

export const Requirement = z.union([
	z.object({
		type: z.literal("purchase_count"),
		amount: z.number(),
	}),
	z.object({
		type: z.literal("email"),
		value: z.string(),
	}),
])

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

export const AdminCoupon = Coupon.extend({
	id: z.number().nullable(),
	stripe_id: z.string(),
	enabled: z.boolean(),
	public: z.boolean(),
	stripe_desc: z.string().nullish(),
});

export type CartItem = z.infer<typeof CartItem>
export type Requirement = z.infer<typeof Requirement>
export type Discount = z.infer<typeof Discount>
export type Coupon = z.infer<typeof Coupon>
export type AdminCoupon = z.infer<typeof AdminCoupon>

export const calculateCartTotal = (cart: Item[]) =>
	cart.reduce<number>((total, item) => total + item.unitPrice * item.amount, 0);

export const applyCoupon = (cart: Item[], coupon: Coupon): number => {
	const cartTotal = calculateCartTotal(cart);
	switch (coupon.discount.type) {
		case 'percentage':
			return Math.floor((cartTotal * (100 - coupon.discount.amount)) / 100);
	}
};

export const checkRequirement = (cart: Item[], email: string, requirement: Requirement): boolean => {
	switch (requirement.type) {
		case 'purchase_count':
			return cart.reduce<number>((total, item) => total + item.amount, 0) >= requirement.amount;
		case 'email':
			return (email + EMAIL_SUFFIX).toLowerCase() === requirement.value.toLowerCase()
	}
};
