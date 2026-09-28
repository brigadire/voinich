"""Isolated inference entry point. Receives neutral surface inputs only."""
import json
import os
from pathlib import Path
import resource
import sys

PACKAGE = Path(__file__).resolve().parent
sys.dont_write_bytecode = True
sys.path.insert(0, str(PACKAGE))


def install_guard(surface, output, checkpoint):
    allowed_read = {surface.resolve(), PACKAGE/'engine.py', PACKAGE/'worker.py'}
    allowed_write = {output.resolve(), output.with_name(output.name+'.tmp').resolve()}
    if checkpoint:
        allowed_read.add(checkpoint.resolve())
        allowed_write.update({checkpoint.resolve(), checkpoint.with_name(checkpoint.name+'.tmp').resolve()})
    runtime = Path(sys.base_prefix).resolve()
    def guard(event, args):
        if event != 'open' or isinstance(args[0], int):
            return
        path = Path(args[0]).resolve()
        mode = args[1] or ''
        flags = args[2] or 0
        writing = any(x in mode for x in 'wax+') or bool(flags & (os.O_WRONLY|os.O_RDWR|os.O_CREAT))
        # Directory fsync opens no contents.
        if path in {output.parent.resolve(), checkpoint.parent.resolve() if checkpoint else output.parent.resolve()} and not writing:
            return
        if writing:
            if path not in allowed_write:
                raise PermissionError('worker write isolation')
        elif path not in allowed_read and not (path.is_relative_to(runtime) and not path.is_relative_to(PACKAGE)):
            raise PermissionError('worker surface-only read isolation')
    sys.addaudithook(guard)


def main():
    surface, output = Path(sys.argv[1]), Path(sys.argv[2])
    checkpoint = Path(sys.argv[3]) if len(sys.argv)>3 and sys.argv[3]!='-' else None
    install_guard(surface, output, checkpoint)
    import engine
    cfg = engine.Config()
    resource.setrlimit(resource.RLIMIT_AS, (cfg.memory_mb*1024*1024, cfg.memory_mb*1024*1024))
    # Hard CPU guard is above wall limit; parent converts abnormal exit to INCOMPLETE.
    resource.setrlimit(resource.RLIMIT_CPU, (int(cfg.seconds*2+10), int(cfg.seconds*2+11)))
    data = json.loads(surface.read_text())
    if set(data) != {'train', 'heldout'}:
        raise ValueError('neutral partitions required')
    for split in data.values():
        if set(split) != {'terms', 'labels'}:
            raise ValueError('surface terms and labels only')
    train = engine.search(data['train']['terms'], data['train']['labels'], cfg, checkpoint,
                          resume='--resume' in sys.argv,
                          pause_after=1 if '--pause' in sys.argv else None)
    test = engine.heldout(train, data['train']['terms'], data['train']['labels'],
                          data['heldout']['terms'], data['heldout']['labels'], cfg)
    engine.atomic_json(output, {'train': train, 'heldout': test, 'accepted': engine.accepted(train, test)})


if __name__ == '__main__':
    main()
