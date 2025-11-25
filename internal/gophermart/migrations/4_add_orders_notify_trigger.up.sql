CREATE OR REPLACE FUNCTION notify_new_order() RETURNS trigger AS $$
DECLARE payload json;
BEGIN payload := json_build_object(
    'order_id',
    NEW.uid,
    'user_id',
    NEW.user_id,
    'number',
    NEW.number,
    'status',
    NEW.status,
    'created_at',
    NEW.uploaded_at
);
PERFORM pg_notify('new_orders', payload::text);
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_notify_new_order ON orders;
CREATE TRIGGER trg_notify_new_order
AFTER
INSERT ON orders FOR EACH ROW
    WHEN (NEW.status = 'NEW') EXECUTE FUNCTION notify_new_order();