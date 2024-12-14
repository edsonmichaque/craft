# tasks.py
from invoke import task, Collection

CI_SCRIPTS = './scripts/ci'

# Test tasks
@task(help={'args': 'Additional arguments for test command'})
def test(ctx, args=''):
    """Run tests"""
    ctx.run(f"{CI_SCRIPTS}/test {args}")

@task
def test_watch(ctx):
    """Run tests in watch mode"""
    ctx.run(f"{CI_SCRIPTS}/test --watch")

@task
def test_coverage(ctx):
    """Run tests with coverage"""
    ctx.run(f"{CI_SCRIPTS}/test coverage")

# Build tasks
@task(help={'args': 'Additional arguments for build command'})
def build(ctx, args=''):
    """Build project"""
    ctx.run(f"{CI_SCRIPTS}/build {args}")

@task
def docker(ctx):
    """Build Docker images"""
    ctx.run(f"{CI_SCRIPTS}/build docker")

# CI tasks
@task
def ci(ctx):
    """Run CI pipeline"""
    ctx.run(f"{CI_SCRIPTS}/ci")

@task
def ci_test(ctx, platform=''):
    """Test CI configurations"""
    cmd = f"{CI_SCRIPTS}/utils/ci-tester.sh"
    if platform:
        cmd += f" {platform}"
    ctx.run(cmd)

# Database tasks
@task
def db_start(ctx):
    """Start database"""
    ctx.run(f"{CI_SCRIPTS}/utils/db.sh start")

@task
def db_migrate(ctx):
    """Run database migrations"""
    ctx.run(f"{CI_SCRIPTS}/utils/db.sh migrate")

@task
def db_seed(ctx):
    """Seed database"""
    ctx.run(f"{CI_SCRIPTS}/utils/db.sh seed")

@task(db_start, db_migrate, db_seed)
def db_reset(ctx):
    """Reset database"""
    print("Database reset complete")

# Release tasks
@task(help={'type': 'Release type (rc, hotfix)'})
def release(ctx, type=''):
    """Create release"""
    type_arg = f"--{type}" if type else ""
    ctx.run(f"{CI_SCRIPTS}/tasks/release.sh {type_arg}")

# Create namespaces
ns = Collection()
ns.add_task(test)
ns.add_task(test_watch)
ns.add_task(test_coverage)
ns.add_task(build)
ns.add_task(docker)
ns.add_task(ci)
ns.add_task(ci_test)

# Database namespace
db = Collection('db')
db.add_task(db_start, 'start')
db.add_task(db_migrate, 'migrate')
db.add_task(db_seed, 'seed')
db.add_task(db_reset, 'reset')
ns.add_collection(db)
