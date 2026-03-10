import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription
} from '~/components/Dialog';

type ModalProps = {
  open: boolean;
  setOpen: (open: boolean) => void;
  children: React.ReactNode;
  title?: string;
  description?: string;
};

export default function Modal(props: ModalProps) {
  const { open, setOpen, children, title, description } = props;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        {title && (
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
            {description && (
              <DialogDescription>{description}</DialogDescription>
            )}
          </DialogHeader>
        )}
        {children}
      </DialogContent>
    </Dialog>
  );
}
