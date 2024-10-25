import requests

def func1(x):
    url = "http://localhost:9090"
    response = requests.get(url)
    return response.json()


if __name__ == "__main__":
    args = list(range(10))
    from multiprocessing import Pool
    with Pool() as mp_pool:
        results = mp_pool.imap(func1, args)
        for result in results:
            print(result)    # func1的返回值