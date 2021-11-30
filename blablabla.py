
kvadratas = [7, 8, 9, 4, 5, 6, 1, 2, 3]
laimejimai = [[True, False, False, True, False, False, True, False, False],
              [False, True, False, False, True, False, False, True, False],
              [False, False, True, False, False, True, False, False, True],
              [True, True, True, False, False, False, False, False, False],
              [False, False, False, True, True, True, False, False, False],
              [False, False, False, False, False, False, True, True, True],
              [True, False, False, False, True, False, False, False, True],
              [False, False, True, False, True, False, True, False, False]]

zaidejas = "X"


def atspausdinti_kvadrata():
    eile = 0
    for simbolis in kvadratas:
        print(str(simbolis) + "|", end="")
        eile += 1
        if eile == 3:
            print()
            eile = 0


def tikrinti_laimejima():
    pakeistas = kvadratas.copy()
    for counter, x in enumerate(pakeistas):
        if x == zaidejas:
            pakeistas[counter] = True
        else:
            pakeistas[counter] = False
    for laimejimas in laimejimai:
        for nr, langelis in enumerate(laimejimas):
            if langelis == True:
                if pakeistas[nr] == langelis:
                    continue
                else:
                    break
        else:
            return True
    return False


def ar_lygiosios():
    for x in kvadratas:
        if type(x) is int:
            return False
    else:
        return True


while True:
    atspausdinti_kvadrata()
    pasirinkimas = int(input(f"Žaidėjas {zaidejas}: pasirinkite veiksmą"))
    if pasirinkimas in kvadratas:
        kvadratas[kvadratas.index(pasirinkimas)] = zaidejas
        if ar_lygiosios():
            print("Lygiosios!")
            break

        if tikrinti_laimejima():
            print(f"Žaidėjas {zaidejas} laimėjo!")
            break

        if zaidejas == "X":
            zaidejas = "O"
        else:
            zaidejas = "X"

    else:
        print("Nėra tokio pasirinkimo, bandykite dar kartą")
