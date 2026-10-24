/** The review queue changed (an item was resolved or discarded): the
    sidebar badge reloads. One shared constant so the event name can
    never drift between dispatcher and listener. */

export const REVIEW_QUEUE_CHANGED_EVENT = 'review-queue-changed';
